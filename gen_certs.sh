keydir="certs"
cd "$keydir"

openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 100000 -out ca.crt -subj "/CN=admission_ca"
openssl genrsa -out warden.key 2048
openssl req -new -key warden.key -out warden.csr -subj "/CN=grumpy.default.svc" -config ../server.conf
openssl x509 -req -in warden.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out warden.crt -days 100000 -extensions v3_req -extfile ../server.conf

# cp warden.crt cert.pem
# cp warden.key key.pem

cp warden.crt grumpy-crt.pem
cp warden.key grumpy-key.pem

keydir="../"
cd "$keydir"

# INJECT CA IN THE WEBHOOK CONFIGURATION
export CA_BUNDLE=$(cat certs/grumpy-crt.pem | base64 | tr -d '\n')
cat _manifest_.yaml | envsubst > manifest.yaml
