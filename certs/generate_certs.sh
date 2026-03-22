#!/bin/bash
set -e

KEY_BITS="${KEY_BITS:-4096}"
KEY_PASSPHRASE="${KEY_PASSPHRASE:-}"

generate_rsa_key() {
	local output_file="$1"
	openssl genpkey -algorithm RSA -pkeyopt "rsa_keygen_bits:${KEY_BITS}" -out "$output_file"
}

write_encrypted_copy() {
	local input_file="$1"
	local output_file="$2"

	if [ -z "$KEY_PASSPHRASE" ]; then
		return
	fi

	KEY_PASSPHRASE="$KEY_PASSPHRASE" openssl pkcs8 -topk8 -inform PEM -outform PEM -v2 aes-256-cbc -in "$input_file" -out "$output_file" -passout env:KEY_PASSPHRASE
}

# Create certs directory if it doesn't exist
mkdir -p certs
cd certs

echo "Generating Root CA..."
# Generate Root CA Key
generate_rsa_key rootCA.key
# Generate Root CA Certificate
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.pem -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Root/CN=GoPDFSuitRootCA"
write_encrypted_copy rootCA.key rootCA.encrypted.key

echo "Generating Intermediate CA..."
# Generate Intermediate Key
generate_rsa_key intermediate.key
# Generate Intermediate CSR
openssl req -new -key intermediate.key -out intermediate.csr -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Intermediate/CN=GoPDFSuitIntermediateCA"
# Generate Intermediate Certificate signed by Root CA
openssl x509 -req -in intermediate.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out intermediate.pem -days 500 -sha256 -extfile <(printf "basicConstraints=CA:TRUE\nkeyUsage=critical,digitalSignature,cRLSign,keyCertSign")
write_encrypted_copy intermediate.key intermediate.encrypted.key

echo "Generating Leaf Certificate..."
# Generate Leaf Key
generate_rsa_key leaf.key
# Generate Leaf CSR
openssl req -new -key leaf.key -out leaf.csr -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Leaf/CN=GoPDFSuitSigner"
# Generate Leaf Certificate signed by Intermediate CA
openssl x509 -req -in leaf.csr -CA intermediate.pem -CAkey intermediate.key -CAcreateserial -out leaf.pem -days 365 -sha256
write_encrypted_copy leaf.key leaf.encrypted.key

echo "Creating Certificate Chain..."
# Create Chain (Intermediate + Root)
cat intermediate.pem rootCA.pem > chain.pem

echo "Done!"
echo "Files created in $(pwd):"
echo "- rootCA.pem (Root Certificate)"
echo "- intermediate.pem (Intermediate Certificate)"
echo "- chain.pem (Intermediate + Root Chain)"
echo "- leaf.pem (Signer Certificate - Use this for 'Certificate PEM')"
echo "- leaf.key (Signer Private Key - Use this for 'Private Key PEM')"
if [ -n "$KEY_PASSPHRASE" ]; then
	echo "- *.encrypted.key (AES-256-CBC encrypted PKCS#8 copies for key-at-rest protection)"
fi
