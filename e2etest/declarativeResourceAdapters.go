// Copyright © Microsoft <wastore@microsoft.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package e2etest

import (
	"github.com/Azure/azure-storage-azcopy/v10/sddl"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/azure-storage-file-go/azfile"
)

func sval(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Our resource adapters convert from objectProperties to the metadata and other objects for Blob/File etc
// They don't share any common interface, because blob/file etc don't share a common interface.
// The reverse conversion, from the remote format back to objectProperties, is in resourceManager.getAllProperties.

// Adapts testObject to blob.
// Doesn't need to deal with anything except contentHeaders and metadata, because that's all Blob supports
type blobResourceAdapter struct {
	obj *testObject
}

func (a blobResourceAdapter) toHeaders() azblob.BlobHTTPHeaders {
	props := a.obj.creationProperties.contentHeaders
	if props == nil {
		return azblob.BlobHTTPHeaders{}
	}
	return azblob.BlobHTTPHeaders{
		ContentType:        sval(props.contentType),
		ContentMD5:         props.contentMD5,
		ContentEncoding:    sval(props.contentEncoding),
		ContentLanguage:    sval(props.contentLanguage),
		ContentDisposition: sval(props.contentDisposition),
		CacheControl:       sval(props.cacheControl),
	}
}

func (a blobResourceAdapter) toMetadata() azblob.Metadata {
	if a.obj.creationProperties.nameValueMetadata == nil {
		return azblob.Metadata{}
	}
	return a.obj.creationProperties.nameValueMetadata
}

func (a blobResourceAdapter) toBlobTags() azblob.BlobTagsMap {
	if a.obj.creationProperties.blobTags == nil {
		return azblob.BlobTagsMap{}
	}
	return azblob.BlobTagsMap(a.obj.creationProperties.blobTags)
}

////

type filesResourceAdapter struct {
	obj *testObject
}

func (a filesResourceAdapter) toHeaders(c asserter, share azfile.ShareURL) azfile.FileHTTPHeaders {
	headers := azfile.FileHTTPHeaders{}

	if a.obj.creationProperties.smbPermissionsSddl != nil {
		parsedSDDL, err := sddl.ParseSDDL(*a.obj.creationProperties.smbPermissionsSddl)
		c.AssertNoErr(err, "Failed to parse SDDL")

		var permKey string

		if len(parsedSDDL.PortableString()) > 8000 {
			createPermResp, err := share.CreatePermission(ctx, parsedSDDL.PortableString())
			c.AssertNoErr(err)

			permKey = createPermResp.FilePermissionKey()
		}

		var smbprops azfile.SMBProperties

		if permKey != "" {
			smbprops.PermissionKey = &permKey
		} else {
			perm := parsedSDDL.PortableString()
			smbprops.PermissionString = &perm
		}

		headers.SMBProperties = smbprops
	}

	if a.obj.creationProperties.smbAttributes != nil {
		attribs := azfile.FileAttributeFlags(*a.obj.creationProperties.smbAttributes)
		headers.SMBProperties.FileAttributes = &attribs
	}

	props := a.obj.creationProperties.contentHeaders
	if props == nil {
		return headers
	}

	headers.ContentType = sval(props.contentType)
	headers.ContentMD5 = props.contentMD5
	headers.ContentEncoding = sval(props.contentEncoding)
	headers.ContentLanguage = sval(props.contentLanguage)
	headers.ContentDisposition = sval(props.contentDisposition)
	headers.CacheControl = sval(props.cacheControl)

	return headers
}

func (a filesResourceAdapter) toMetadata() azfile.Metadata {
	if a.obj.creationProperties.nameValueMetadata == nil {
		return azfile.Metadata{}
	}
	return a.obj.creationProperties.nameValueMetadata
}
