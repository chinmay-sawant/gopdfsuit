#!/bin/bash
set -e

# Create certs directory if it doesn't exist
mkdir -p certs
cd certs

echo "Generating Root CA..."
# Generate Root CA Key
openssl genrsa -out rootCA.key 2048
# Generate Root CA Certificate
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.pem -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Root/CN=GoPDFSuitRootCA"

echo "Generating Intermediate CA..."
# Generate Intermediate Key
openssl genrsa -out intermediate.key 2048
# Generate Intermediate CSR
openssl req -new -key intermediate.key -out intermediate.csr -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Intermediate/CN=GoPDFSuitIntermediateCA"
# Generate Intermediate Certificate signed by Root CA
openssl x509 -req -in intermediate.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out intermediate.pem -days 500 -sha256 -extfile <(printf "basicConstraints=CA:TRUE\nkeyUsage=critical,digitalSignature,cRLSign,keyCertSign")

echo "Generating Leaf Certificate..."
# Generate Leaf Key
openssl genrsa -out leaf.key 2048
# Generate Leaf CSR
openssl req -new -key leaf.key -out leaf.csr -subj "/C=US/ST=State/L=City/O=GoPDFSuit/OU=Leaf/CN=GoPDFSuitSigner"
# Generate Leaf Certificate signed by Intermediate CA
openssl x509 -req -in leaf.csr -CA intermediate.pem -CAkey intermediate.key -CAcreateserial -out leaf.pem -days 365 -sha256

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
