package remote

import (
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// CheckPushPermission returns an error if the given keychain cannot authorize
// a push operation to the given ref.
//
// This can be useful to check whether the caller has permission to push an
// image before doing work to construct the image.
func CheckPushPermission(ref name.Reference, kc authn.Keychain, t http.RoundTripper) error {
	auth, err := kc.Resolve(ref.Context().Registry)
	if err != nil {
		return err
	}

	scopes := []string{ref.Scope(transport.PushScope)}
	tr, err := transport.New(ref.Context().Registry, auth, t, scopes)
	if err != nil {
		return err
	}
	// TODO(jasonhall): Against GCR, just doing the token handshake is
	// enough, but this doesn't extend to Dockerhub, so we actually need
	// to initiate an upload. Figure out how to return early here when we
	// can.
	w := writer{
		ref:    ref,
		client: &http.Client{Transport: tr},
	}
	_, _, err = w.initiateUpload("", "")
	return err
}
