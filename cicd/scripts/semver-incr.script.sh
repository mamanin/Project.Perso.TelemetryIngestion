#!/bin/bash
# ─────────────────────────────────────────────────────────────────────────────
# semver-incr.script.sh
#
# Auto-increments a SemVer version stored in an Azure DevOps variable group
# and writes the new value back to the library.
#
# Usage:
#   semver-incr.script.sh \
#     <ORGANIZATION_URL> \
#     <PROJECT_NAME> \
#     <VARIABLE_GROUP> \
#     <VERSION_KEY> \
#     <VERSION_INCREMENT> \
#     <ENABLE_SUFFIX> \
#     <SUFFIX>
#
# Arguments:
#   ORGANIZATION_URL  - ADO org URL, e.g. https://dev.azure.com/myorg
#   PROJECT_NAME      - ADO project name, e.g. MyProject
#   VARIABLE_GROUP    - Name of the ADO variable group (library)
#   VERSION_KEY       - Name of the variable that holds the version
#   VERSION_INCREMENT - Increment strategy: patch | minor | major
#   ENABLE_SUFFIX     - Append a pre-release suffix: true | false
#   SUFFIX            - Suffix string, e.g. dev (only used when ENABLE_SUFFIX=true)
#
# Environment:
#   SYSTEM_ACCESSTOKEN - Azure Pipelines OAuth token (set via env: in the job)
#
# Output:
#   Sets the pipeline variable $(IncrementVersion.version) via logging command.
#   Export as pipeline output variable so subsequent jobs can consume it via:
#     dependencies.VersionIncrement.outputs['IncrementVersion.version']
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

ORGANIZATION_URL="${1:?'Argument 1 (ORGANIZATION_URL) is required'}"
PROJECT_NAME="${2:?'Argument 2 (PROJECT_NAME) is required'}"
VARIABLE_GROUP="${3:?'Argument 3 (VARIABLE_GROUP) is required'}"
VERSION_KEY="${4:?'Argument 4 (VERSION_KEY) is required'}"
VERSION_INCREMENT="${5:?'Argument 5 (VERSION_INCREMENT: patch|minor|major) is required'}"
ENABLE_SUFFIX="${6:-false}"
SUFFIX="${7:-dev}"

ORGANIZATION_URL="${ORGANIZATION_URL%/}"
PROJECT_NAME_ENCODED=$(printf '%s' "$PROJECT_NAME" | jq -sRr @uri)

ACCESS_TOKEN="${SYSTEM_ACCESSTOKEN:?'Environment variable SYSTEM_ACCESSTOKEN is required'}"
BASE64_TOKEN=$(printf ":%s" "$ACCESS_TOKEN" | base64 -w 0)
AUTH_HEADER="Authorization: Basic ${BASE64_TOKEN}"
API_VERSION="7.1"

echo "──────────────────────────────────────────────────────"
echo "  Version Increment"
echo "──────────────────────────────────────────────────────"
echo "  Organization  : ${ORGANIZATION_URL}"
echo "  Project       : ${PROJECT_NAME}"
echo "  Variable Group: ${VARIABLE_GROUP}"
echo "  Version Key   : ${VERSION_KEY}"
echo "  Increment     : ${VERSION_INCREMENT}"
echo "  Suffix        : ${ENABLE_SUFFIX} (value: '${SUFFIX}')"
echo "──────────────────────────────────────────────────────"

echo "[1/5] Fetching variable group '${VARIABLE_GROUP}'..."

GROUPS_URL="${ORGANIZATION_URL}/${PROJECT_NAME_ENCODED}/_apis/distributedtask/variablegroups?groupName=${VARIABLE_GROUP}&api-version=${API_VERSION}"
GROUPS_RESPONSE=$(curl -sf \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  "$GROUPS_URL")

GROUP_COUNT=$(echo "$GROUPS_RESPONSE" | jq -r '.count')
if [[ "$GROUP_COUNT" -eq 0 ]]; then
  echo "ERROR: Variable group '${VARIABLE_GROUP}' was not found in project '${PROJECT_NAME}'."
  exit 1
fi

GROUP_ID=$(echo "$GROUPS_RESPONSE" | jq -r '.value[0].id')
CURRENT_VERSION=$(echo "$GROUPS_RESPONSE" | jq -r --arg key "$VERSION_KEY" '.value[0].variables[$key].value // empty')

if [[ -z "$CURRENT_VERSION" ]]; then
  echo "ERROR: Variable '${VERSION_KEY}' was not found in group '${VARIABLE_GROUP}'."
  exit 1
fi

echo "  Found group ID    : ${GROUP_ID}"
echo "  Current version   : ${CURRENT_VERSION}"

echo "[2/5] Parsing current version..."

VERSION_CORE=$(echo "$CURRENT_VERSION" | sed 's/-.*$//')
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_CORE"

if ! [[ "$MAJOR" =~ ^[0-9]+$ && "$MINOR" =~ ^[0-9]+$ && "$PATCH" =~ ^[0-9]+$ ]]; then
  echo "ERROR: '${CURRENT_VERSION}' (core: '${VERSION_CORE}') is not a valid SemVer (major.minor.patch)."
  exit 1
fi

echo "  Parsed → ${MAJOR}.${MINOR}.${PATCH}"

echo "[3/5] Incrementing version (${VERSION_INCREMENT})..."

case "${VERSION_INCREMENT,,}" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "ERROR: Invalid VERSION_INCREMENT '${VERSION_INCREMENT}'. Allowed values: patch | minor | major."
    exit 1
    ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"

if [[ "${ENABLE_SUFFIX,,}" == "true" && -n "$SUFFIX" ]]; then
  NEW_VERSION="${NEW_VERSION}-${SUFFIX}"
fi

echo "  New version → ${NEW_VERSION}"

echo "[4/5] Fetching full group object for update..."

GROUP_URL="${ORGANIZATION_URL}/${PROJECT_NAME_ENCODED}/_apis/distributedtask/variablegroups/${GROUP_ID}?api-version=${API_VERSION}"

GROUP_OBJECT=$(curl -sf \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  "$GROUP_URL")

# Patch the target variable value inside the JSON object
UPDATED_GROUP=$(echo "$GROUP_OBJECT" | \
  jq --arg key "$VERSION_KEY" --arg value "$NEW_VERSION" \
  '.variables[$key].value = $value')

echo "[5/5] Writing new version to variable group..."

HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -X PUT \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d "$UPDATED_GROUP" \
  "$GROUP_URL")

if [[ "$HTTP_STATUS" -lt 200 || "$HTTP_STATUS" -ge 300 ]]; then
  echo "ERROR: Failed to update variable group '${VARIABLE_GROUP}' (HTTP ${HTTP_STATUS})."
  exit 1
fi

echo "──────────────────────────────────────────────────────"
echo "  Version successfully updated: ${CURRENT_VERSION} → ${NEW_VERSION}"
echo "──────────────────────────────────────────────────────"

echo "##vso[task.setvariable variable=version;isOutput=true]${NEW_VERSION}"
