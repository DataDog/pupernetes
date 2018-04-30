#!/usr/bin/env python
import json
import sys
import urllib
import yaml

if __name__ == '__main__':
    y = yaml.load(sys.stdin.read())
    for elt in y.get("storage", {}).get("files", []):
        try:
            elt["contents"]["source"] = "data:,%s" % urllib.quote(elt["contents"]["inline"], safe='')
            del elt["contents"]["inline"]
        except KeyError:
            pass

    print(json.dumps(y, indent=2))
