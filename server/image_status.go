package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/containers/storage"
	"github.com/cri-o/cri-o/internal/log"
	pkgstorage "github.com/cri-o/cri-o/internal/storage"
	json "github.com/json-iterator/go"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	types "k8s.io/cri-api/pkg/apis/runtime/v1"
)

// ImageStatus returns the status of the image.
func (s *Server) ImageStatus(ctx context.Context, req *types.ImageStatusRequest) (*types.ImageStatusResponse, error) {
	ctx, span := log.StartSpan(ctx)
	defer span.End()
	img := req.Image
	if img == nil || img.Image == "" {
		return nil, fmt.Errorf("no image specified")
	}

	log.Infof(ctx, "Checking image status: %s", img.Image)
	status, err := s.storageImageStatus(ctx, *img)
	if err != nil {
		return nil, err
	}
	if status == nil {
		log.Infof(ctx, "Image %s not found", img.Image)
		return &types.ImageStatusResponse{}, nil
	}

	// Ensure that size is already defined
	var size uint64
	if status.Size == nil {
		size = 0
	} else {
		size = *status.Size
	}

	resp := &types.ImageStatusResponse{
		Image: &types.Image{
			Id:          status.ID,
			RepoTags:    status.RepoTags,
			RepoDigests: status.RepoDigests,
			Size_:       size,
			Spec: &types.ImageSpec{
				Annotations: status.Annotations,
			},
		},
	}
	if req.Verbose {
		info, err := createImageInfo(status)
		if err != nil {
			return nil, fmt.Errorf("creating image info: %w", err)
		}
		resp.Info = info
	}
	uid, username := getUserFromImage(status.User)
	if uid != nil {
		resp.Image.Uid = &types.Int64Value{Value: *uid}
	}
	resp.Image.Username = username
	log.Infof(ctx, "Image status: %v", resp)
	return resp, nil
}

// storageImageStatus calls ImageStatus for a k8s ImageSpec.
// Returns (nil, nil) if image was not found.
func (s *Server) storageImageStatus(ctx context.Context, spec types.ImageSpec) (*pkgstorage.ImageResult, error) {
	images, err := s.StorageImageServer().ResolveNames(s.config.SystemContext, spec.Image)
	if err != nil {
		return nil, err
	}
	var (
		status  *pkgstorage.ImageResult
		lastErr error
	)
	for _, image := range images {
		status_, err := s.StorageImageServer().ImageStatus(s.config.SystemContext, image)
		if err != nil {
			if errors.Is(err, storage.ErrImageUnknown) {
				log.Debugf(ctx, "Can't find %s", image)
				continue
			}
			log.Warnf(ctx, "Error getting status from %s: %v", image, err)
			lastErr = err
			continue
		}
		status = status_
		break
	}
	if status == nil {
		if lastErr != nil {
			return nil, lastErr
		}
		// ResolveNames returns at least one value if it doesn't fail.
		// So, if we got here, there was at least one ErrImageUnknown, and no other errors.
		log.Infof(ctx, "Image %s not found", spec.Image)
		return nil, nil
	}
	return status, nil
}

// getUserFromImage gets uid or user name of the image user.
// If user is numeric, it will be treated as uid; or else, it is treated as user name.
func getUserFromImage(user string) (id *int64, username string) {
	// return both empty if user is not specified in the image.
	if user == "" {
		return nil, ""
	}
	// split instances where the id may contain user:group
	user = strings.Split(user, ":")[0]
	// user could be either uid or user name. Try to interpret as numeric uid.
	uid, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		// If user is non numeric, assume it's user name.
		return nil, user
	}
	// If user is a numeric uid.
	return &uid, ""
}

func createImageInfo(result *pkgstorage.ImageResult) (map[string]string, error) {
	info := struct {
		Labels    map[string]string `json:"labels,omitempty"`
		ImageSpec *specs.Image      `json:"imageSpec"`
	}{
		result.Labels,
		result.OCIConfig,
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %v: %w", info, err)
	}
	return map[string]string{"info": string(bytes)}, nil
}
