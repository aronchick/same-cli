export PRIVATE_KEY_PASSPHRASE="AAAAAAAAAA"

openssl genrsa -aes128 -passout pass:$PRIVATE_KEY_PASSPHRASE -out private.pem 4096
openssl rsa -in private.pem -passin pass:$PRIVATE_KEY_PASSPHRASE -pubout -out public.pem

These then need to appear in the Github secrets - 
PRIVATE_KEY_PASSPHRASE
PRIVATE_PEM
PUBLIC_PEM

You also need to generate a Github token. Put it in a secret named:
SAME_CLI_TESTER_GITHUB_TOKEN