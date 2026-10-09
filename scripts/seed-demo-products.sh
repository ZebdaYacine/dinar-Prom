#!/usr/bin/env bash
set -Eeuo pipefail

# Download public demo product photos, upload them to MinIO, and
# upsert the matching fake product/media rows in PostgreSQL.
# Usage: ENV_FILE=.env.example ./scripts/seed-demo-products.sh

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
ENV_FILE="${ENV_FILE:-.env.example}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
MANIFEST="${MANIFEST:-apps/api/internal/guest/data/demo-products.json}"
CATALOG="${CATALOG:-apps/api/internal/guest/data/consumer-catalog.json}"
PUBLIC_MEDIA_URL="${PUBLIC_MEDIA_URL:-http://localhost:9100/dinar-public}"
MC_IMAGE="minio/mc:RELEASE.2025-04-16T18-13-26Z"

[[ -f "$ENV_FILE" ]] || { echo "Environment file not found: $ENV_FILE" >&2; exit 1; }
[[ -f "$MANIFEST" ]] || { echo "Manifest not found: $MANIFEST" >&2; exit 1; }
[[ -f "$CATALOG" ]] || { echo "Catalog not found: $CATALOG" >&2; exit 1; }
command -v curl >/dev/null || { echo 'curl is required.' >&2; exit 1; }
command -v jq >/dev/null || { echo 'jq is required.' >&2; exit 1; }
command -v docker >/dev/null || { echo 'Docker is required.' >&2; exit 1; }

set -a
source "$ENV_FILE"
set +a

compose() {
  SHARED_ENV_FILE="$ENV_FILE" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

compose config --quiet
compose up -d postgres minio minio-init >/dev/null
NETWORK="$(docker inspect -f '{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{end}}' "$(compose ps -q minio)")"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo 'Downloading demo product photos...'
while IFS= read -r product; do
  id="$(jq -r '.id' <<<"$product")"
  filename="$(basename "$(jq -r '.object_key' <<<"$product")")"
  curl --fail --location --retry 3 --retry-delay 1 --silent --show-error "$(jq -r '.source_url' <<<"$product")" -o "$TMP_DIR/$filename"
done < <(jq -c '.[]' "$MANIFEST")

echo 'Uploading product photos to MinIO...'
docker run --rm --network "$NETWORK" -e MINIO_ROOT_USER -e MINIO_ROOT_PASSWORD -v "$TMP_DIR:/seed:ro" --entrypoint /bin/sh "$MC_IMAGE" -ec '
  mc alias set local http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null
  mc mb --ignore-existing local/dinar-public >/dev/null
  mc anonymous set download local/dinar-public >/dev/null
  mc rm --recursive --force local/dinar-public/demo-products >/dev/null || true
  for file in /seed/*; do name="$(basename "$file")"; mc cp --attr "Content-Type=image/jpeg" "$file" "local/dinar-public/demo-products/$name" >/dev/null; done
'

echo 'Applying demo product schema...'
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" < apps/api/db/migrations/000002_demo_products.up.sql >/dev/null
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" < apps/api/db/migrations/000003_catalog_brands.up.sql >/dev/null
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" < apps/api/db/migrations/000004_seed_catalog_brands.up.sql >/dev/null
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" < apps/api/db/migrations/000005_consumer_catalog.up.sql >/dev/null
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" < apps/api/db/migrations/000006_consumer_audience.up.sql >/dev/null
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" < apps/api/db/migrations/000007_brand_origin.up.sql >/dev/null

sql_escape() { printf '%s' "$1" | sed "s/'/''/g"; }

echo 'Seeding consumer categories, subcategories, and global/local brands...'
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "DELETE FROM subcategory_brands WHERE category_slug IN ('fashion','beauty','phones','computers','food','furniture','kitchenware','appliances'); DELETE FROM catalog_subcategories; DELETE FROM catalog_categories;" >/dev/null
# Remove only the legacy taxonomy namespaces created by the original demo seed.
# Consumer categories above are rebuilt from the JSON catalog on every run.
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "DELETE FROM subcategory_brands WHERE category_slug IN ('cars','electronics'); DELETE FROM brands WHERE NOT EXISTS (SELECT 1 FROM subcategory_brands WHERE subcategory_brands.brand_id = brands.id);" >/dev/null

category_order=0
while IFS= read -r category; do
  category_slug="$(jq -r '.slug' <<<"$category")"
  category_name_ar="$(sql_escape "$(jq -r '.name_ar' <<<"$category")")"
  category_name_fr="$(sql_escape "$(jq -r '.name_fr' <<<"$category")")"
  category_icon="$(sql_escape "$(jq -r '.icon' <<<"$category")")"
  compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "INSERT INTO catalog_categories (slug,name_ar,name_fr,icon,sort_order) VALUES ('$(sql_escape "$category_slug")','$category_name_ar','$category_name_fr','$category_icon',$category_order) ON CONFLICT (slug) DO UPDATE SET name_ar=EXCLUDED.name_ar,name_fr=EXCLUDED.name_fr,icon=EXCLUDED.icon,sort_order=EXCLUDED.sort_order,updated_at=now();" </dev/null >/dev/null
  sub_order=0
  while IFS= read -r subcategory; do
    subcategory_slug="$(jq -r '.slug' <<<"$subcategory")"
    subcategory_name_ar="$(sql_escape "$(jq -r '.name_ar' <<<"$subcategory")")"
    subcategory_name_fr="$(sql_escape "$(jq -r '.name_fr' <<<"$subcategory")")"
    compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "INSERT INTO catalog_subcategories (slug,category_slug,name_ar,name_fr,sort_order) VALUES ('$(sql_escape "$subcategory_slug")','$(sql_escape "$category_slug")','$subcategory_name_ar','$subcategory_name_fr',$sub_order) ON CONFLICT (slug) DO UPDATE SET category_slug=EXCLUDED.category_slug,name_ar=EXCLUDED.name_ar,name_fr=EXCLUDED.name_fr,sort_order=EXCLUDED.sort_order,updated_at=now();" </dev/null >/dev/null
    sub_order=$((sub_order + 1))
  done < <(jq -c '.subcategories[]' <<<"$category")
  category_order=$((category_order + 1))
done < <(jq -c '.categories[]' "$CATALOG")

brand_order=0
while IFS= read -r brand; do
  brand_slug="$(sql_escape "$(jq -r '.slug' <<<"$brand")")"
  brand_name_ar="$(sql_escape "$(jq -r '.name_ar' <<<"$brand")")"
  brand_name_fr="$(sql_escape "$(jq -r '.name_fr' <<<"$brand")")"
  brand_origin="$(sql_escape "$(jq -r '.origin // "global"' <<<"$brand")")"
  brand_category="$(sql_escape "$(jq -r '.category_id' <<<"$brand")")"
  brand_subcategory_id="$(jq -r '.subcategory_id' <<<"$brand")"
  brand_subcategory="$(sql_escape "$(jq -r --arg id "$brand_subcategory_id" '.categories[].subcategories[] | select(.id == $id) | .slug' "$CATALOG")")"
  compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "INSERT INTO brands (slug,name_ar,name_fr,origin) VALUES ('$brand_slug','$brand_name_ar','$brand_name_fr','$brand_origin') ON CONFLICT (slug) DO UPDATE SET name_ar=EXCLUDED.name_ar,name_fr=EXCLUDED.name_fr,origin=EXCLUDED.origin,updated_at=now(); INSERT INTO subcategory_brands (brand_id,category_slug,subcategory_slug,sort_order) SELECT id,'$brand_category','$brand_subcategory',$brand_order FROM brands WHERE slug='$brand_slug' ON CONFLICT (brand_id,category_slug,subcategory_slug) DO UPDATE SET sort_order=EXCLUDED.sort_order;" </dev/null >/dev/null
  brand_order=$((brand_order + 1))
done < <(jq -c '.brands[]' "$CATALOG")
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c 'DELETE FROM brands WHERE NOT EXISTS (SELECT 1 FROM subcategory_brands WHERE subcategory_brands.brand_id = brands.id);' >/dev/null

echo 'Linking MinIO objects to fake PostgreSQL product rows...'
compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c 'TRUNCATE TABLE demo_product_media, demo_products RESTART IDENTITY;' >/dev/null
while IFS= read -r product; do
  id="$(jq -r '.id' <<<"$product")"
  title_ar="$(sql_escape "$(jq -r '.title_ar' <<<"$product")")"
  title_fr="$(sql_escape "$(jq -r '.title_fr' <<<"$product")")"
  description_ar="$(sql_escape "$(jq -r '.description_ar' <<<"$product")")"
  description_fr="$(sql_escape "$(jq -r '.description_fr' <<<"$product")")"
  category="$(jq -r '.category' <<<"$product")"
  subcategory="$(jq -r '.subcategory' <<<"$product")"
  brand_slug="$(jq -r '.brand // empty' <<<"$product")"
  audience_gender="$(jq -r '.audience_gender // "unisex"' <<<"$product")"
  age_min="$(jq -r '.age_min // 0' <<<"$product")"
  age_max="$(jq -r '.age_max // 120' <<<"$product")"
  wilaya_code="$(jq -r '.wilaya_code' <<<"$product")"
  city="$(sql_escape "$(jq -r '.city' <<<"$product")")"
  city_ar="$(sql_escape "$(jq -r '.city_ar' <<<"$product")")"
  object_key="$(sql_escape "$(jq -r '.object_key' <<<"$product")")"
  source_url="$(sql_escape "$(jq -r '.source_url' <<<"$product")")"
  filename="$(basename "$(jq -r '.object_key' <<<"$product")")"
  byte_size="$(wc -c < "$TMP_DIR/$filename")"
  public_url="$(sql_escape "$PUBLIC_MEDIA_URL/$(jq -r '.object_key' <<<"$product")")"
  brand_value='NULL'
  if [[ -n "$brand_slug" ]]; then brand_value="'$(sql_escape "$brand_slug")'"; fi
  compose exec -T postgres psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "INSERT INTO demo_products (external_id,title_ar,title_fr,description_ar,description_fr,category_slug,subcategory_slug,brand_slug,audience_gender,age_min,age_max,wilaya_code,city_name,city_name_ar,source_url,status) VALUES ('$id','$title_ar','$title_fr','$description_ar','$description_fr','$category','$subcategory',$brand_value,'$(sql_escape "$audience_gender")',$age_min,$age_max,'$wilaya_code','$city','$city_ar','$source_url','approved') ON CONFLICT (external_id) DO UPDATE SET title_ar=EXCLUDED.title_ar,title_fr=EXCLUDED.title_fr,description_ar=EXCLUDED.description_ar,description_fr=EXCLUDED.description_fr,category_slug=EXCLUDED.category_slug,subcategory_slug=EXCLUDED.subcategory_slug,brand_slug=EXCLUDED.brand_slug,audience_gender=EXCLUDED.audience_gender,age_min=EXCLUDED.age_min,age_max=EXCLUDED.age_max,wilaya_code=EXCLUDED.wilaya_code,city_name=EXCLUDED.city_name,city_name_ar=EXCLUDED.city_name_ar,source_url=EXCLUDED.source_url,updated_at=now(); INSERT INTO demo_product_media (product_id,object_key,bucket_name,source_url,public_url,content_type,byte_size,position) SELECT id,'$object_key','dinar-public','$source_url','$public_url','image/jpeg',$byte_size,1 FROM demo_products WHERE external_id='$id' ON CONFLICT (object_key) DO UPDATE SET product_id=EXCLUDED.product_id,source_url=EXCLUDED.source_url,public_url=EXCLUDED.public_url,byte_size=EXCLUDED.byte_size;" </dev/null >/dev/null
done < <(jq -c '.[]' "$MANIFEST")

count="$(compose exec -T postgres psql -At -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c 'SELECT count(*) FROM demo_products;')"
media_count="$(compose exec -T postgres psql -At -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c 'SELECT count(*) FROM demo_product_media;')"
printf 'Seed completed: %s products, %s linked MinIO media rows.\n' "$count" "$media_count"
printf 'Media base URL: %s\n' "$PUBLIC_MEDIA_URL"
