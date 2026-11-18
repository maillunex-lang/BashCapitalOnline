#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 ]]; then
  printf 'Usage: %s preview|release /absolute/path/to/output.apk\n' "$0" >&2
  exit 2
fi

root=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
output=$(realpath -m "$2")
if [[ "$output" != *.apk ]]; then
  printf 'The output filename must end in .apk\n' >&2
  exit 2
fi
mkdir -p "$(dirname "$output")"

case "$1" in
  preview)
    export_mode=--export-debug
    preset="Android Preview"
    package=io.github.maillunexlang.bashcapital.preview
    ;;
  release)
    : "${GODOT_ANDROID_KEYSTORE_RELEASE_PATH:?Set the authorized release keystore path}"
    : "${GODOT_ANDROID_KEYSTORE_RELEASE_USER:?Set the release keystore alias}"
    : "${GODOT_ANDROID_KEYSTORE_RELEASE_PASSWORD:?Set the release keystore password}"
    export_mode=--export-release
    preset="Android Release"
    package=io.github.maillunexlang.bashcapital
    ;;
  *) printf 'Unknown build channel: %s\n' "$1" >&2; exit 2 ;;
esac

: "${ANDROID_HOME:?Set ANDROID_HOME to the Android SDK path}"
build_tools="$ANDROID_HOME/build-tools/${ANDROID_BUILD_TOOLS_VERSION:-35.0.1}"
godot_bin="${GODOT_BIN:-godot}"

git -C "$root" diff --quiet HEAD -- || {
  printf 'Commit the tracked source before producing a build with provenance.\n' >&2
  exit 1
}
commit=$(git -C "$root" rev-parse HEAD)
engine=$("$godot_bin" --version)
scratch=$(mktemp -d)
trap 'python3 -c "import shutil,sys; shutil.rmtree(sys.argv[1])" "$scratch"' EXIT
apk="$scratch/$(basename "$output")"

"$godot_bin" --headless --path "$root/client" --editor --import
"$godot_bin" --headless --path "$root/client" "$export_mode" "$preset" "$apk"
"$build_tools/apksigner" verify --verbose --print-certs "$apk" | tee "$scratch/signature.txt"
"$build_tools/aapt" dump badging "$apk" > "$scratch/badging.txt"
grep -q "^package: name='$package'" "$scratch/badging.txt"
grep -q "^uses-permission: name='android.permission.INTERNET'" "$scratch/badging.txt"
"$build_tools/aapt" dump xmltree "$apk" AndroidManifest.xml > "$scratch/manifest.txt"
grep -q 'android.intent.action.MAIN' "$scratch/manifest.txt"
grep -q 'android.intent.category.LAUNCHER' "$scratch/manifest.txt"
if [[ "$1" == release ]] && grep -q '^application-debuggable' "$scratch/badging.txt"; then
  printf 'Release export is debuggable.\n' >&2
  exit 1
fi
unzip -tq "$apk"
unzip -Z1 "$apk" > "$scratch/entries.txt"
grep -qx 'assets/project.binary' "$scratch/entries.txt"
grep -qx 'assets/scripts/main.gdc' "$scratch/entries.txt"
grep -qx 'lib/arm64-v8a/libgodot_android.so' "$scratch/entries.txt"
grep -qx 'lib/x86_64/libgodot_android.so' "$scratch/entries.txt"
if grep -q '^assets/tests/' "$scratch/entries.txt"; then
  printf 'Test scripts should be excluded from the APK.\n' >&2
  exit 1
fi

mv "$apk" "$output"
(cd "$(dirname "$output")" && sha256sum "$(basename "$output")") > "$output.sha256"
{
  printf 'Git commit: %s\nGodot: %s\nChannel: %s\nServer economy catalog tested separately: 2\n' "$commit" "$engine" "$1"
  grep -E "^(package:|sdkVersion:|targetSdkVersion:|native-code:)" "$scratch/badging.txt"
  cat "$scratch/signature.txt" "$output.sha256"
} > "$output.build.txt"
printf 'Verified APK: %s\n' "$output"
