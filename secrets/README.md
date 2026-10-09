# Runtime secrets

Generate JWT keys here for local/production deployment, but never commit the PEM
files. The Compose files mount these files as Docker secrets.

```bash
mkdir -p secrets
openssl genrsa -out secrets/jwt_private.pem 4096
openssl rsa -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem
chmod 600 secrets/jwt_private.pem
chmod 644 secrets/jwt_public.pem
```
