from PIL import Image
from tqdm import tqdm
import cv2
import os
import sys


if __name__ == "__main__":
    completed = "stitched_results"

    tiles_path = f"{sys.argv[1]}"
    lod = sys.argv[2]

    tiles_in_dir = sorted(os.listdir(tiles_path))
    r = 0
    c = 0

    for t in tiles_in_dir[::-1]:
        if t.startswith(lod):
            sp = t.split("_")
            r = max(r, int(sp[1]))
            c = max(c, int(sp[2].split(".")[0]))

    r2b = lambda rgb: (cv2.imread(rgb))
    verticals = []

    ext = tiles_in_dir[1].split(".")[1]
    for ridx in tqdm(range(r + 1)):
        outvec = []
        for cidx in range((c + 1)):
            tilepath = f"{os.path.join(tiles_path)}/{lod}_{ridx}_{cidx}.{ext}"
            outvec.append(r2b(tilepath))

        filename = os.path.join("tmp", f"{lod}_{ridx}.{ext}")

        verticals.append(cv2.hconcat(outvec))

    # write out our completed image
    vstack = cv2.vconcat(verticals)

    outname = tiles_path.split("/")[-1]
    resname = os.path.join(completed, f"{lod}_{outname}.{ext}")

    cv2.imwrite(f"{resname}", vstack)

    exit(0)
