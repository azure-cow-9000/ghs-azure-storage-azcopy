// Copyright © 2017 Microsoft <wastore@microsoft.com>
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

package ste

import (
	"strings"
	"time"

	"github.com/Azure/azure-storage-azcopy/v10/azbfs"
	"github.com/Azure/azure-storage-azcopy/v10/common"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// Source info provider for Azure blob
type blobSourceInfoProvider struct {
	defaultRemoteSourceInfoProvider
}

func newBlobSourceInfoProvider(jptm IJobPartTransferMgr) (ISourceInfoProvider, error) {
	base, err := newDefaultRemoteSourceInfoProvider(jptm)
	if err != nil {
		return nil, err
	}

	return &blobSourceInfoProvider{defaultRemoteSourceInfoProvider: *base}, nil
}

// AccessControl should ONLY get called when we know for a fact it is a blobFS->blobFS tranfser.
// It *assumes* that the source is actually a HNS account.
func (p *blobSourceInfoProvider) AccessControl() (azbfs.BlobFSAccessControl, error) {
	presignedURL, err := p.PreSignedSourceURL()
	if err != nil {
		return azbfs.BlobFSAccessControl{}, err
	}

	bURLParts := azblob.NewBlobURLParts(*presignedURL)
	bURLParts.Host = strings.ReplaceAll(bURLParts.Host, ".blob", ".dfs")
	bURLParts.BlobName = strings.TrimSuffix(bURLParts.BlobName, "/") // BlobFS doesn't handle folders correctly like this.
	// todo: jank, and violates the principle of interfaces
	fURL := azbfs.NewFileURL(bURLParts.URL(), p.jptm.(*jobPartTransferMgr).jobPartMgr.(*jobPartMgr).secondarySourceProviderPipeline)

	return fURL.GetAccessControl(p.jptm.Context())
}

func (p *blobSourceInfoProvider) BlobTier() azblob.AccessTierType {
	return p.transferInfo.S2SSrcBlobTier
}

func (p *blobSourceInfoProvider) BlobType() azblob.BlobType {
	return p.transferInfo.SrcBlobType
}

func (p *blobSourceInfoProvider) GetFreshFileLastModifiedTime() (time.Time, error) {
	presignedURL, err := p.PreSignedSourceURL()
	if err != nil {
		return time.Time{}, err
	}

	blobURL := azblob.NewBlobURL(*presignedURL, p.jptm.SourceProviderPipeline())
	clientProvidedKey := azblob.ClientProvidedKeyOptions{}
	if p.jptm.IsSourceEncrypted() {
		clientProvidedKey = common.ToClientProvidedKeyOptions(p.jptm.CpkInfo(), p.jptm.CpkScopeInfo())
	}

	properties, err := blobURL.GetProperties(p.jptm.Context(), azblob.BlobAccessConditions{}, clientProvidedKey)
	if err != nil {
		return time.Time{}, err
	}

	return properties.LastModified(), nil
}
