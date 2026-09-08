---
---

https://ceur-ws.org/Vol-4064/PD-paper9.pdf

```
mkdir -p in out
```

```
mkdir -p in/schema/v29.3
curl -Lo in/schema/v29.3/source.ttl https://raw.githubusercontent.com/schemaorg/schemaorg/refs/tags/v29.3-release/data/schema.ttl
cat <<EOF > in/schema/v29.3/config.properties
codeRepository=https://github.com/schemaorg/schemaorg
ontologyName=schema.org
ontologyPrefix=schema
ontologyNamespaceURI=https://schema.org/
previousVersionURI=https://raw.githubusercontent.com/schemaorg/schemaorg/refs/tags/v29.2-release/data/schema.ttl
source=https://github.com/schemaorg/schemaorg/blob/v29.3-release/data/schema.ttl
thisVersionURI=https://raw.githubusercontent.com/schemaorg/schemaorg/refs/tags/v29.3-release/data/schema.ttl
EOF
podman run -ti --rm \
  --volume "${PWD}/in:/usr/local/widoco/in:Z" \
  --volume "${PWD}/out:/usr/local/widoco/out:Z" \
  --platform linux/amd64 \
  ghcr.io/dgarijo/widoco:v1.4.25 \
    -confFile in/schema/v29.3/config.properties \
    -ontFile in/schema/v29.3/source.ttl \
    -outFolder out/schema/v29.3 \
    -includeAnnotationProperties \
    -rewriteAll
```

```
mkdir -p in/schema/main
curl -Lo in/schema/main/source.ttl https://raw.githubusercontent.com/schemaorg/schemaorg/refs/heads/main/data/schema.ttl
cat <<EOF > in/schema/main/config.properties
codeRepository=https://github.com/schemaorg/schemaorg
ontologyName=schema.org
ontologyPrefix=schema
ontologyNamespaceURI=https://schema.org/
previousVersionURI=https://raw.githubusercontent.com/schemaorg/schemaorg/refs/tags/v29.3-release/data/schema.ttl
source=https://github.com/schemaorg/schemaorg/blob/main-release/data/schema.ttl
thisVersionURI=https://raw.githubusercontent.com/schemaorg/schemaorg/refs/heads/main/data/schema.ttl
EOF
podman run -ti --rm \
  --volume "${PWD}/in:/usr/local/widoco/in:Z" \
  --volume "${PWD}/out:/usr/local/widoco/out:Z" \
  --platform linux/amd64 \
  ghcr.io/dgarijo/widoco:v1.4.25 \
    -confFile in/schema/main/config.properties \
    -ontFile in/schema/main/source.ttl \
    -outFolder out/schema/main \
    -includeAnnotationProperties \
    -rewriteAll
```

```
curl -Lo in/schema-main.ttl https://raw.githubusercontent.com/schemaorg/schemaorg/refs/heads/main/data/schema.ttl
podman run -ti --rm \
  --volume "${PWD}/in:/usr/local/widoco/in:Z" \
  --volume "${PWD}/out:/usr/local/widoco/out:Z" \
  --platform linux/amd64 \
  ghcr.io/dgarijo/widoco:v1.4.25 \
    -ontFile in/schema-main.ttl \
    -outFolder out/schema-main \
    -rewriteAll
```
