package signature

import (
	"encoding/asn1"
	"time"
)

// P42: hand-built DER for PKCS#7 SignedData assembly. Static signer fields are
// pre-encoded at NewPDFSigner; per-sign work patches digest, UTCTime, and signature.

type pkcs7DERParts struct {
	contentTypeAttr []byte // SEQUENCE: contentType + data OID
	digestAlgID     []byte // AlgorithmIdentifier SHA-256
	sigAlgID        []byte // AlgorithmIdentifier ECDSA/RSA
	digestAlgsSET   []byte // SET { digestAlgID }
	encapContent    []byte // SEQUENCE { data OID }
	issuerSerial    []byte // IssuerAndSerialNumber
	certsContext0   []byte // [0] IMPLICIT certificates
}

func marshalAlgorithmIdentifier(oid asn1.ObjectIdentifier) []byte {
	b, err := asn1.Marshal(pkixAlgorithmIdentifier{Algorithm: oid})
	if err != nil {
		panic(err)
	}
	return b
}

func buildPKCS7DERParts(s *PDFSigner) pkcs7DERParts {
	contentTypeOID := mustMarshal(oidContentType)
	dataOIDSET := mustMarshal(asn1.RawValue{
		Class:      asn1.ClassUniversal,
		Tag:        asn1.TagSet,
		IsCompound: true,
		Bytes:      marshaledOIDData,
	})
	attr1Content := contentTypeOID
	attr1Content = append(attr1Content, dataOIDSET...)
	contentTypeAttr := wrapDERTag(asn1.TagSequence, attr1Content)

	digestAlgID := marshalAlgorithmIdentifier(oidSHA256)
	digestAlgsSET := wrapDERTag(asn1.TagSet, digestAlgID)
	encapContent := wrapDERTag(asn1.TagSequence, mustMarshal(oidData))

	issuerSerial, err := asn1.Marshal(s.issuerAndSerial)
	if err != nil {
		panic(err)
	}

	certsContext0 := wrapDERContext0(s.precomputedCertBytes)

	return pkcs7DERParts{
		contentTypeAttr: contentTypeAttr,
		digestAlgID:     digestAlgID,
		sigAlgID:        marshalAlgorithmIdentifier(s.digestSigAlgorithm),
		digestAlgsSET:   digestAlgsSET,
		encapContent:    encapContent,
		issuerSerial:    issuerSerial,
		certsContext0:   certsContext0,
	}
}

func appendDERLength(dst []byte, n int) []byte {
	if n < 0x80 {
		return append(dst, byte(n))
	}
	switch {
	case n <= 0xff:
		return append(dst, 0x81, byte(n))
	case n <= 0xffff:
		return append(dst, 0x82, byte(n>>8), byte(n))
	default:
		return append(dst, 0x83, byte(n>>16), byte(n>>8), byte(n))
	}
}

func wrapDERTag(tag byte, content []byte) []byte {
	out := make([]byte, 0, 2+len(content))
	out = append(out, tag)
	out = appendDERLength(out, len(content))
	return append(out, content...)
}

func wrapDERContext0(content []byte) []byte {
	out := make([]byte, 0, 2+len(content))
	out = append(out, 0xa0)
	out = appendDERLength(out, len(content))
	return append(out, content...)
}

func appendUTCTime(dst []byte, t time.Time) []byte {
	utc := t.UTC().Format("060102150405") + "Z"
	dst = append(dst, 0x17)
	dst = appendDERLength(dst, len(utc))
	return append(dst, utc...)
}

func buildMessageDigestAttr(messageDigest []byte) []byte {
	digestOID := mustMarshal(oidMessageDigest)
	octet := wrapDERTag(asn1.TagOctetString, messageDigest)
	digestSET := wrapDERTag(asn1.TagSet, octet)
	attrContent := digestOID
	attrContent = append(attrContent, digestSET...)
	return wrapDERTag(asn1.TagSequence, attrContent)
}

func buildSigningTimeAttr(signingTime time.Time) []byte {
	timeOID := mustMarshal(oidSigningTime)
	var utcBuf [32]byte
	utc := appendUTCTime(utcBuf[:0], signingTime)
	timeSET := wrapDERTag(asn1.TagSet, utc)
	attrContent := timeOID
	attrContent = append(attrContent, timeSET...)
	return wrapDERTag(asn1.TagSequence, attrContent)
}

func buildAuthenticatedAttributesSET(parts pkcs7DERParts, messageDigest []byte, signingTime time.Time) []byte {
	attr2 := buildMessageDigestAttr(messageDigest)
	attr3 := buildSigningTimeAttr(signingTime)
	content := make([]byte, 0, len(parts.contentTypeAttr)+len(attr2)+len(attr3))
	content = append(content, parts.contentTypeAttr...)
	content = append(content, attr2...)
	content = append(content, attr3...)
	return wrapDERTag(asn1.TagSet, content)
}

func stripOuterTLV(der []byte) []byte {
	if len(der) < 2 {
		return der
	}
	offset := 1
	if der[offset]&0x80 == 0 {
		offset++
	} else {
		numBytes := int(der[offset] & 0x7f)
		offset += 1 + numBytes
	}
	if offset > len(der) {
		return nil
	}
	return der[offset:]
}

func buildSignerInfoDER(parts pkcs7DERParts, authAttrsContent []byte, signature []byte) []byte {
	version := []byte{0x02, 0x01, 0x01}
	authAttrs := wrapDERContext0(authAttrsContent)
	sigOctet := wrapDERTag(asn1.TagOctetString, signature)

	content := make([]byte, 0, 256)
	content = append(content, version...)
	content = append(content, parts.issuerSerial...)
	content = append(content, parts.digestAlgID...)
	content = append(content, authAttrs...)
	content = append(content, parts.sigAlgID...)
	content = append(content, sigOctet...)
	return wrapDERTag(asn1.TagSequence, content)
}

func buildSignedDataDER(parts pkcs7DERParts, signerInfo []byte) []byte {
	version := []byte{0x02, 0x01, 0x01}
	signerInfosSET := wrapDERTag(asn1.TagSet, signerInfo)

	content := make([]byte, 0, len(parts.certsContext0)+len(signerInfosSET)+64)
	content = append(content, version...)
	content = append(content, parts.digestAlgsSET...)
	content = append(content, parts.encapContent...)
	content = append(content, parts.certsContext0...)
	content = append(content, signerInfosSET...)
	return wrapDERTag(asn1.TagSequence, content)
}

func buildContentInfoDER(signedData []byte) []byte {
	oidBytes := mustMarshal(oidSignedData)
	content := wrapDERContext0(signedData)
	outer := make([]byte, 0, len(oidBytes)+len(content))
	outer = append(outer, oidBytes...)
	outer = append(outer, content...)
	return wrapDERTag(asn1.TagSequence, outer)
}
