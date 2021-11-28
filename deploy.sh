docker build . -t rewanthtammana/grumpy
docker push rewanthtammana/grumpy

kubectl delete secret grumpy

./gen_certs.sh

kubectl create secret generic grumpy -n default \
	  --from-file=key.pem=certs/grumpy-key.pem \
	    --from-file=cert.pem=certs/grumpy-crt.pem

kubectl apply -f manifest.yaml

# Copy notary certificates!
