# Prep the release of vestaluna
# Involves using the Fyne packaging system, and makes usage of zip
#!/usr/bin/python

# usage: `python prep_release.py windows` #other options are mac, linux
#

import os
import datetime
import sys

target = sys.argv[1]
rel_or_debug = sys.argv[2]

if __name__ == "__main__":
    print("Building app...")

    match target:
        ## WINDOWS ##
        case "windows"  or "win":
            os.system(f"fyne-cross {target} -release -icon assets/icons/vestaluna_logo.png")
            print("Copying assets...")
            os.system("cp -r scripts fyne-cross/bin/windows-amd64")
            os.system("cp -r apiEndPoints.txt fyne-cross/bin/windows-amd64")

            print("Zipping...")
            release_stamp = datetime.datetime().timestamp()
            os.system(f"zip -r {release_stamp}_win_vestaluna.zip fyne-cross/bin/windows-amd64")
        
        ## MAC DARWIN ##
        case "mac"
            os.system(f"fyne-cross {target} -ap-id vestaluna -release -icon assets/icons/vestaluna_logo.png")
            print("Copying assets...")
            os.system("cp -r scripts fyne-cross/bin/darwin-amd64")
            os.system("cp -r apiEndPoints.txt fyne-cross/bin/darwin-amd64")

            print("Zipping...")
            release_stamp = datetime.datetime().timestamp()
            os.system(f"zip -r {release_stamp}_darwin_vestaluna.zip fyne-cross/bin/darwin-amd64")

        ## LINUX ##
        case _:
            os.system(f"fyne-cross {target} -release -icon assets/icons/vestaluna_logo.png")
            print("Copying assets...")
            os.system("cp -r scripts fyne-cross/bin/linux-amd64")
            os.system("cp -r apiEndPoints.txt fyne-cross/bin/linux-amd64")

            print("Zipping...")
            release_stamp = datetime.datetime().timestamp()
            os.system(f"zip -r {release_stamp}_linux_vestaluna.zip fyne-cross/bin/windows-amd64")


    print("Cleaning up...")
    os.system("rm -rf fyne-cross")
