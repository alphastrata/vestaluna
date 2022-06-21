from PIL import Image
from tqdm import tqdm
import cv2
import numpy as np
import os
import sys


def build_from_stack(stack, filename):
    return cv2.hconcat(stack)


def check(im):
    return im.shape[0] == im.shape[1]


if __name__ == "__main__":
    # Get tiles_path and LOD from sysargs

    completed = "stitched_results"

    tiles_path = f"downloads/{target}/{sys.argv[1] if len(sys.argv) > 1 else exit("No tile_path supplied.")}"
    lod = sys.argv[2] if len(sys.argv) > 2 else exit("No level of detail supplied.")
    ext = (
        sys.argv[3]
        if len(sys.argv) > 3
        else exit("No image format supplied for tiles.")
    )

    r2b = lambda rgb: (cv2.imread(rgb))
    verticals = []

    for ridx in tqdm(range(cols)):
        outvec = []
        for cidx in range(len(rows)):
        # Sample filename: 3_6_15.png
            tilepath = f"{os.path.join(tiles_path)}/{lod}_{ridx}_{cidx}.{ext}"
            print(tilepath)
            img = r2b(os.path.join(tiles_path, tilepath)
            outvec.append(img)

        filename = f"{os.path.join("tmp"), row_idx}.jpg" # NOTE: design descision to convert to jpg here...

        verticals.append(build_from_stack(outvec, filename))

    # write out our completed image
    vstack = cv2.vconcat(verticals)

    result = os.path.join(completed, f"{hhmmss}.png"), vstack
    cv2.imwrite(result)

    cv2.imshow(result)
    if cv2.waitKey(5) & 0xFF == 27:
        break
        cv2.destroyAllWindows()
