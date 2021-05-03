export PRIVATE_KEY_PASSPHRASE="AAAAAAAAAA"
gh secret set PRIVATE_KEY_PASSPHRASE -b"${PRIVATE_KEY_PASSPHRASE}"

echo $PRIVATE_KEY_PASSPHRASE | gh secret set PRIVATE_KEY_PASSPHRASE

openssl genrsa -aes128 -passout pass:$PRIVATE_KEY_PASSPHRASE -out /tmp/private.pem 4096

gh secret set PRIVATE_PEM < /tmp/private.pem

openssl rsa -in /tmp/private.pem -passin pass:$PRIVATE_KEY_PASSPHRASE -pubout -out /tmp/public.pem

gh secret set PUBLIC_PEM <  /tmp/public.pem

You then need to go to the (SAME-Project/SAME-installer-website)[https://github.com/SAME-Project/SAME-installer-website/blob/main/install_script.sh#L32] and update the public.pem with the output of /tmp/public.pem.

Don't forget to delete these files!

You also need to generate a Github token. Put it in a secret named:
SAME_CLI_TESTER_GITHUB_TOKEN
