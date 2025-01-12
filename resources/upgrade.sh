
#!/usr/bin/env bash


distroRoot="/opt"
imagesRoot="$distroRoot/images"

load_containerd_images(){
   cd   $imagesRoot/cri_images_update
   gzip --decompress *.tar.gz
   for f in *
   do
        ctr -n=k8s.io images import $f
        echo "successfully imported $f to containerd"
   done
}

load_containerd_images