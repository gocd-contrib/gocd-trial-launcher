#!/bin/bash

set -e

PROGNAME="run-gocd"

rm -f "$PROGNAME"
rm -rf dist meta

RELEASE="X.x.x"

for arg in $@; do
  case $arg in
    --verbose)
      extra_flags="$extra_flags -v"
      shift
      ;;
    --skip-tests)
      skip=true
      shift
      ;;
    --prod)
      multiplatform=true
      shift
      ;;
    --release)
      RELEASE="$2"
      shift
      ;;
    --release=*)
      RELEASE="${arg#*=}"
      shift
      ;;
    *)
      shift
      ;;
  esac
done

RELEASE="${RELEASE}-${GO_PIPELINE_LABEL:-localbuild}"

function ldflags {
  local _os="${1:-$(go env GOOS)}"
  local _arch="${2:-$(go env GOARCH)}"

  local flags="-X main.Version=${RELEASE} -X main.GitCommit=${GIT_COMMIT} -X main.Platform=${_arch}-${_os}"

  # embed Info.plist
  if [ "darwin" = "$_os" ]; then
    flags="${flags} -linkmode external -extldflags \"-sectcreate __TEXT __info_plist meta/Info.plist\""
  fi

  echo "$flags"
}

echo "Fetching dependencies"
go get -d $extra_flags ./...

echo "Fetching any windows-specific dependencies"
GOOS="windows" go get -d $extra_flags ./... # get any windows-specific deps as well

if [[ "true" = "$skip" ]]; then
  echo "Skipping tests"
else
   go test $extra_flags ./...
fi

if (which git &> /dev/null); then
  GIT_COMMIT=$(git rev-list --abbrev-commit -1 HEAD)
else
  GIT_COMMIT="unknown"
fi

mkdir -p meta
cat <<INFOPLIST > meta/Info.plist
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<plist version="1.0">
  <dict>
    <key>CFBundleName</key>
    <string>${PROGNAME}</string>
    <key>CFBundleShortVersionString</key>
    <string>${RELEASE}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleDevelopmentRegion</key>
    <string>English</string>
    <key>CFBundleVersion</key>
    <string>${RELEASE}</string>
    <key>CFBundleIdentifier</key>
    <string>org.gocd.testdrive.launcher</string>
    <key>NSHumanReadableCopyright</key>
    <string>${PROGNAME} ${RELEASE}. Copyright ThoughtWorks Inc., (c) 2000-2019</string>
  </dict>
</plist>
INFOPLIST

if [[ "true" = "$multiplatform" ]]; then
  platforms=(
    darwin/amd64
    linux/amd64
    windows/amd64
  )

  echo "Release: ${RELEASE}, Revision: ${GIT_COMMIT}"

  for plt in "${platforms[@]}"; do
    mkdir -p "dist/${plt}"
    arr=(${plt//\// })
    _os="${arr[0]}"
    _arch="${arr[1]}"
    name="$PROGNAME"

    if [[ "windows" = "${_os}" ]]; then
      name="$name.exe"
    fi

    echo "Building $plt..."

    GOOS="${_os}" GOARCH="${_arch}" go build \
      -o "dist/${plt}/${name}" \
      -ldflags "$(ldflags "$_os" "$_arch")" \
      main.go
  done
else
  go build \
    -ldflags "$(ldflags)" \
    -o "$PROGNAME" \
    main.go
fi
