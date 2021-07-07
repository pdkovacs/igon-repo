#!/bin/bash

###############################################################################
# "Triages" some of icons at https://github.com/google/material-design-icons
# for testing/demo purposes.
#
# Note for Mac OS:
#   brew install coreutils
#   alias readlink=greadlink
#
###############################################################################

#
# The following characteristics of the extracted set can be observed:
#
# Raster
# - formats: png
# - sizes: 18dp, 24dp, 36dp, 48dp, 144dp
#
# Svg:
# - sizes: 18px, 24px, 48px
#

path_to_google_matdes_icon_repo=${HOME}/github/google/material-design-icons
destination_root=$(readlink -f $(dirname $0)/../test/demo-data)
selector_name_length_modulus=2


rm -rf "$destination_root"
mkdir $destination_root

echo "destination_root: $destination_root"

cd "$path_to_google_matdes_icon_repo"

declare -A icon_count
file_count=0

filter_by_icon_name_length_mod() {
    (( ${#icon_name} % $selector_name_length_modulus == 0 ))
}

store() {
    f=$1 size=$2; format=$3; icon_name=$4;

    filter_by_icon_name_length_mod || return

    destdir="$destination_root/$format/$size"
set -x
    mkdir -p "$destdir"
    cp $f $destdir/$icon_name.$format
    icon_count["$icon_name"]=$((icon_count["$icon_name"]+1))
set +x
    file_count=$((file_count+1))
}

raster() {
    while read f;
    do
        icon_name=$(echo "$f" | sed -e 's/.*\/ic_\([^.]\+\)_black_[^.]\+\.[^.]\+/\1/g')
        format=$(echo "$f" | sed -e 's/.*\/ic_[^.]\+_black_[^.]\+\.\([^.]\+\)/\1/g')
        size=$(echo "$f" | sed -e 's/.*\/ic_[^.]\+_black_\([^.]\+\)\.[^.]\+/\1/g')
        store $f $size $format $icon_name
    done < <(find . -type f | \
                egrep '(^\./editor/.*$|^\./hardware/.*$)' | \
                grep -v /ios/ | \
                grep drawable-hdpi | \
                egrep '*_black_[^.]+.[a-z]+' | \
                sort)
}

svg() {
    while read f;
    do
        icon_name=$(echo "$f" | sed -e 's/.*\/ic_\([^.]\+\)_[^._]\+\.svg/\1/g')
        format=svg
        size=$(echo "$f" | sed -e 's/.*\/ic_[^.]\+_\([^._]\+\)\.svg/\1/g')
        store $f $size $format $icon_name
    done < <(find . -type f | \
                egrep '(^\./editor/.*$|^\./hardware/.*$)' | \
                grep -v /ios/ | \
                grep svg | \
                sort)
}

raster
svg

cd "$destination_root"
du -h .
echo "Stored ${#icon_count[@]} icon(s) in $file_count file(s)"
