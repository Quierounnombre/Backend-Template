openssl genrsa 2048 | tee private.pem
openssl rsa -in private.pem -pubout | tee public.pem

awk 'NF {printf "%s\\n", $0}' private.pem > private_env.pem
awk 'NF {printf "%s\\n", $0}' public.pem > public_env.pem

rm private.pem
rm public.pem
