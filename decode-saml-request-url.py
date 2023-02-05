#!/bin/python3
from urllib.parse import urlparse, parse_qs
from base64 import b64decode
import zlib

# NB you should replace this value with your own.
url = 'https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/saml2?SAMLRequest=TODO'

u = urlparse(url)

qs = parse_qs(u.query)

saml_request_deflated_base64 = qs['SAMLRequest'][0]

saml_request_deflated = b64decode(saml_request_deflated_base64)

saml_request = zlib.decompress(saml_request_deflated, -15)

print(saml_request.decode('utf-8'))
