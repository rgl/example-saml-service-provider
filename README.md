# About

This is an example SAML Service Provider.

# Usage (SAMLtest IdP)

Build and run in foreground:

```bash
make
```

Open the [testing samltest.id IdP page](https://samltest.id) at:

https://samltest.id/upload.php

And upload the `example-saml-service-provider-metadata.xml` file (you can see
it at http://localhost:8000/saml/metadata too).

Open this example Service Provider page, and click the `login` link to go
tru the authentication flow:

http://localhost:8000

# Usage (Azure AD IdP)

**NB** You can also [use terraform to automatically create and configure](https://github.com/rgl/example-saml-service-provider-azure) the resources described in this section.

Open the [Azure Portal](https://portal.azure.com).

Select the `Azure Active Directory` resource.

Click the `Manage`, `Enterprise applications` menu.

Click the `+ New application` button.

Click the `+ Create your own application` button.

Create the application.

Open the created application.

Click the `Manage`, `Users and groups` menu.

Click the `+ Add user/group` button and assign the users who can use the application.

Click the `Manage`, `Single sign-on` menu.

Click the `Edit` button under the `Basic SAML Configuration` section, and set the following properties:

| Property                                   | Value                                       |
|--------------------------------------------|---------------------------------------------|
| Identifier (Entity ID)                     | `urn:example:example-saml-service-provider` |
| Reply URL (Assertion Consumer Service URL) | `http://localhost:8000/saml/acs`            |

Copy the application `App Federation Metadata Url` property value to the
`EXAMPLE_IDP_METADATA` environment variable, e.g.:

```bash
export EXAMPLE_IDP_METADATA='https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/federationmetadata/2007-06/federationmetadata.xml?appid=1bc7df9a-9a80-4c3a-9a2b-6f737b7d0a70'
```

Build and run this example Service Provider in foreground:

```bash
make build
./example-saml-service-provider \
    --idp-metadata $EXAMPLE_IDP_METADATA
```

Open this example Service Provider page, and click the `login` link to go
tru the authentication flow:

http://localhost:8000

You should see a list of SAML Attributes/Claims. Something like:

| Name                                                                  | Value                                                                                 |
|-----------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| `http://schemas.microsoft.com/claims/authnmethodsreferences`          | `http://schemas.microsoft.com/ws/2008/06/identity/authenticationmethod/password`      |
| `http://schemas.microsoft.com/claims/authnmethodsreferences`          | `http://schemas.microsoft.com/claims/multipleauthn`                                   |
| `http://schemas.microsoft.com/claims/authnmethodsreferences`          | `http://schemas.microsoft.com/ws/2008/06/identity/authenticationmethod/unspecified`   |
| `http://schemas.microsoft.com/identity/claims/displayname`            | `Rui Lopes`                                                                           |
| `http://schemas.microsoft.com/identity/claims/identityprovider`       | `live.com`                                                                            |
| `http://schemas.microsoft.com/identity/claims/objectidentifier`       | `00000000-0000-0000-0000-000000000000`                                                |
| `http://schemas.microsoft.com/identity/claims/tenantid`               | `00000000-0000-0000-0000-000000000000`                                                |
| `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress`  | `rui@example.com`                                                                     |
| `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname`     | `Rui`                                                                                 |
| `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name`          | `rui_example.com#EXT#@example.onmicrosoft.com`                                        |
| `http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname`       | `Lopes`                                                                               |
| `SessionIndex`                                                        | `_00000000-0000-0000-0000-000000000000`                                               |

# Troubleshoot

* To debug a SAML request inside a URL redirect, edit the `url` property inside
  the `decode-saml-request-url.py` file and execute it to see the SAML request
  XML document.
  * The `SAMLRequest` query string value is encoded as
    `zlib.deflate(base64encode(saml_request_xml_document))`.
