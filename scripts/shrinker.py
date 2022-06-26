# resize all the images in the completed/ folder to 25% of their original size using opencv

import os
import cv2
from tqdm import tqdm

complete_discs = os.listdir(os.path.join(os.getcwd(), "../stitched_results"))

for i in tqdm(complete_discs):
    img = cv2.imread(os.path.join(os.getcwd(), "../stitched_results", i))
    img = cv2.resize(img, (0, 0), fx=0.15, fy=0.15)
    i = i.replace(".png", "_15percent_.png")
    cv2.imwrite(os.path.join(os.getcwd(), "../stitched_results", i), img)
