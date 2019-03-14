#!/bin/bash

set -e

GOCD_JRE_URL="${GOCD_JRE_URL:-https://s3.amazonaws.com/mirrors-archive/local/jdk}"
GOCD_JRE_VERSION="${GOCD_JRE_VERSION:-11.0.2}"

SCRATCH_DIR="scratch"
INSTALLERS_DIR="installers"

# The main entry point; takes an arbitrary list of platforms for which
# to assemble installers.
function main {
  echo "Creating installers for platforms: $@..."

  rm -rf "$INSTALLERS_DIR" "$SCRATCH_DIR" # clean up from prior runs
  mkdir -p "deps"

  echo "Checking for required artifacts..."

  # First, resolve the server zip archive
  local server_zip=$(find deps/zip -type f -name 'go-server-*.zip' -print | head -n 1)

  if [ -z "$server_zip" -o ! -s "$server_zip" ]; then
    die "Cannot find the go-server zip artifact. Aborting."
  fi

  # Parse the version + build from the server zip archive
  local GOCD_VERSION=$(parse_version_from_server_zip "$server_zip")

  # Resolve the corresponding agent (i.e., same version + build)
  local agent_zip="deps/zip/go-agent-${GOCD_VERSION}.zip"

  if [ -z "$agent_zip" -o ! -s "$agent_zip" ]; then
    die "Cannot find the go-agent zip artifact. Aborting."
  fi

  echo "Building installer version ${GOCD_VERSION} bundled with JRE ${GOCD_JRE_VERSION}..."

  # Extract server and agent contents
  unpack_server "$server_zip" "${SCRATCH_DIR}/jars"
  unpack_agent "$agent_zip" "${SCRATCH_DIR}/jars"

  echo ""

  for plt in $@; do
    echo "Assembling installer for platform: ${plt}"

    local dest_dir="${SCRATCH_DIR}/installers/${plt}/gocd-${GOCD_VERSION}"
    mkdir -p "${dest_dir}/gocd/packages"

    fetch_jre "$plt"
    prepare_jre "$plt" "${dest_dir}/gocd/packages/jre"
    prepare_server "${dest_dir}/gocd/packages/go-server"
    prepare_agent "${dest_dir}/gocd/packages/go-agent"

    prepare_launcher "$plt" "${dest_dir}/gocd"

    package_installer "$dest_dir" "$INSTALLERS_DIR"

    echo "Done."
    echo ""
  done

  echo "Cleaning up scratch direcory..."

  rm -rf "$SCRATCH_DIR" # cleanup only on success; we otherwise want to inspect the working dir

  echo "Success."
}

# Creates a zip archive from SRC at DEST
# Usage:
#  package_installer SRC DEST
function package_installer {
  local name="$(basename "$1")"
  local src_dir="$(dirname "$1")"
  local plt="$(basename "$src_dir")"
  local wd="$(pwd)"
  local archive_name="$2/${name}-${plt}.zip"

  mkdir -p "$(dirname "${wd}/${archive_name}")"

  echo "  * Packaging ${archive_name}... [src: $1]"

  (cd "$src_dir" && zip -qr "${wd}/${archive_name}" "${name}")
}

# Resolves the correct launcher for the specified OS/platform
# and puts it into the respective assembly folder
function prepare_launcher {
  local plt="$1"
  local dest="$2"

  echo "  * Bundling run-gocd launcher..."

  case "$plt" in
    osx)
      local src="dist/darwin/amd64/run-gocd"
      ;;
    linux)
      local src="dist/linux/amd64/run-gocd"
      ;;
    windows)
      local src="dist/windows/amd64/run-gocd.exe"
      ;;
    *)
      die "Don't know path to run-gocd for platform: ${plt}"
      ;;
  esac

  chmod a+rx "$src" # when getting artifact via the fetchArtifact task, it may lose the executable flag

  if [ ! -f "$src" -o ! -s "$src" ]; then
    die "The \`run-gocd\` binary is missing; expected to find it at: ${src}"
  fi

  ln -f "$src" "${dest}/"

  local top_level_dir="$(dirname "$dest")"

  case "$plt" in
    windows)
      local wrapper="${top_level_dir}/run-me.cmd"
      cat <<WINCMD > "$wrapper"
cmd /c gocd\run-gocd.exe

WINCMD
      ;;
    *)
      local wrapper="${top_level_dir}/run-me.sh"
      cat <<NIXCMD > "$wrapper"
#!/bin/bash

set -e

cd "\$(dirname "\$0")"
chmod a+rx gocd/run-gocd

gocd/run-gocd

NIXCMD
      ;;
  esac

  chmod a+rx "$wrapper"

  if [ "osx" = "$plt" ]; then
    mv "$wrapper" "${top_level_dir}/run-me.command"
  fi
}

# Unpacks the downloaded OS-specific JDK, trims it down to a JRE, and puts
# it into the correct assembly folder
function prepare_jre {
  local plt="$1"
  local dest="$2"

  echo "  * Unpacking and preparing JRE for ${plt}..."

  local workdir="${SCRATCH_DIR}/${plt}"
  mkdir -p "$workdir"

  unpack_to "deps/jdk/$(jre_pkg_name "$plt")" "$workdir"

  if [ "$plt" = osx ]; then
    local src="${workdir}/jdk-${GOCD_JRE_VERSION}.jdk/Contents/Home"
  else
    local src="${workdir}/jdk-${GOCD_JRE_VERSION}"
  fi

  mv "${src}" "$dest"
  rm -rf "$workdir"

  rm -rf "${dest}/lib/src.zip" "${dest}/jmods" "${dest}/include"
}

# Puts the server uber-jar into the assembly folder
function prepare_server {
  local dest="$1"
  mkdir -p "$dest"

  echo "  * Bundling server dependency..."

  # hard-link so we don't need to extract the jar for each platform
  ln "${SCRATCH_DIR}/jars/go-server/go.jar" "$dest/"
}

# Puts the agent uber-jar into the assembly folder
function prepare_agent {
  local dest="$1"
  mkdir -p "$dest"

  echo "  * Bundling agent dependency..."

  # hard-link so we don't need to extract the jar for each platform
  ln "${SCRATCH_DIR}/jars/go-agent/agent-bootstrapper.jar" "$dest/"
}

# Downloads the JDK package for the specified OS/platform
function fetch_jre {
  local plt="$1"
  local dest="deps/jdk"
  local jre_pkg=$(jre_pkg_name "$plt")

  mkdir -p "$dest"

  echo "  * Fetching JRE for ${plt}..."

  if [ ! -f "${dest}/${jre_pkg}" -o ! -s "${dest}/${jre_pkg}" ]; then # prevent unnecessary downloads during dev
    local jre_url="${GOCD_JRE_URL}/${jre_pkg}"
    echo "    * Using url: ${jre_url}"

    curl "$jre_url" -o "${dest}/${jre_pkg}"
  fi
}

function parse_version_from_server_zip {
  local server_zip="$1"
  local _base=$(basename "$server_zip" .zip)
  echo "${_base#"go-server-"}"
}

function unpack_server {
  local archive="$1"
  local dest="$2"
  local version_only="${GOCD_VERSION%-*}" # strip off build number

  echo "Unpacking GoCD server go.jar..."

  unpack_to "$archive" "$dest" "go-server-${version_only}/go.jar"
  mv "${dest}/go-server-${version_only}" "${dest}/go-server"
}

function unpack_agent {
  local archive="$1"
  local dest="$2"
  local version_only="${GOCD_VERSION%-*}" # strip off build number

  echo "Unpacking GoCD agent agent-bootstrapper.jar..."

  unpack_to "$archive" "$dest" "go-agent-${version_only}/agent-bootstrapper.jar"
  mv "${dest}/go-agent-${version_only}" "${dest}/go-agent"
}

function die {
  echo $@ >&2
  exit 1
}

function jre_pkg_name {
  local plt="$1"
  local base_name="openjdk-${GOCD_JRE_VERSION}_${plt}-x64_bin"

  if [ "$plt" = "windows" ]; then
    echo "${base_name}.zip"
  else
    echo "${base_name}.tar.gz"
  fi
}

# Unpacks an archive to a specified directory; smart enough to unpack both
# *.zip and *.tar.gz
#
# Destination directory is automatically created it does not exist.
#
# Usage:
#   unpack_to ARCHIVE DEST_DIR [ FILE_1 ] ... [ FILE_N ]
#
# Where FILE_1..N are optional args to allow extracting only specified files
# from ARCHIVE
function unpack_to {
  local src="$1"
  local dst="$2"

  shift 2

  mkdir -p "$dst"

  case "$src" in
    *.zip)
      unzip -qo "$src" $@ -d "$dst"
      ;;
    *.tar.gz)
      tar zxf "$src" -C "$dst" $@
      ;;
    *)
      die "Don't know how to unpack archive: $src"
      ;;
  esac
}

main $@
