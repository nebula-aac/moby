package layer

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/containerd/log"
	"github.com/docker/distribution"
	"github.com/moby/moby/v2/pkg/ioutils"
	"github.com/moby/sys/atomicwriter"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

var supportedAlgorithms = []digest.Algorithm{
	digest.SHA256,
	// digest.SHA384, // Currently not used
	// digest.SHA512, // Currently not used
}

type fileMetadataStore struct {
	root string
}

type fileMetadataTransaction struct {
	store *fileMetadataStore
	ws    *atomicwriter.WriteSet
}

// newFSMetadataStore returns an instance of a metadata store
// which is backed by files on disk using the provided root
// as the root of metadata files.
func newFSMetadataStore(root string) (*fileMetadataStore, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	return &fileMetadataStore{
		root: root,
	}, nil
}

func (fms *fileMetadataStore) getLayerDirectory(layer ChainID) string {
	return filepath.Join(fms.root, string(layer.Algorithm()), layer.Encoded())
}

func (fms *fileMetadataStore) getLayerFilename(layer ChainID, filename string) string {
	return filepath.Join(fms.getLayerDirectory(layer), filename)
}

func (fms *fileMetadataStore) getMountDirectory(mount string) string {
	return filepath.Join(fms.root, "mounts", mount)
}

func (fms *fileMetadataStore) getMountFilename(mount, filename string) string {
	return filepath.Join(fms.getMountDirectory(mount), filename)
}

func (fms *fileMetadataStore) StartTransaction() (*fileMetadataTransaction, error) {
	tmpDir := filepath.Join(fms.root, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, err
	}
	ws, err := atomicwriter.NewWriteSet(tmpDir)
	if err != nil {
		return nil, err
	}

	return &fileMetadataTransaction{
		store: fms,
		ws:    ws,
	}, nil
}

func (fm *fileMetadataTransaction) SetSize(size int64) error {
	return fm.ws.WriteFile("size", []byte(strconv.FormatInt(size, 10)), 0o644)
}

func (fm *fileMetadataTransaction) SetParent(parent ChainID) error {
	return fm.ws.WriteFile("parent", []byte(parent.String()), 0o644)
}

func (fm *fileMetadataTransaction) SetDiffID(diff DiffID) error {
	return fm.ws.WriteFile("diff", []byte(diff.String()), 0o644)
}

func (fm *fileMetadataTransaction) SetCacheID(cacheID string) error {
	return fm.ws.WriteFile("cache-id", []byte(cacheID), 0o644)
}

func (fm *fileMetadataTransaction) SetDescriptor(ref distribution.Descriptor) error {
	jsonRef, err := json.Marshal(ref)
	if err != nil {
		return err
	}
	return fm.ws.WriteFile("descriptor.json", jsonRef, 0o644)
}

func (fm *fileMetadataTransaction) TarSplitWriter(compressInput bool) (io.WriteCloser, error) {
	f, err := fm.ws.FileWriter("tar-split.json.gz", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	var wc io.WriteCloser
	if compressInput {
		wc = gzip.NewWriter(f)
	} else {
		wc = f
	}

	return ioutils.NewWriteCloserWrapper(wc, func() error {
		wc.Close()
		return f.Close()
	}), nil
}

func (fm *fileMetadataTransaction) Commit(layer ChainID) error {
	finalDir := fm.store.getLayerDirectory(layer)
	if err := os.MkdirAll(filepath.Dir(finalDir), 0o755); err != nil {
		return err
	}

	return fm.ws.Commit(finalDir)
}

func (fm *fileMetadataTransaction) Cancel() error {
	return fm.ws.Cancel()
}

func (fm *fileMetadataTransaction) String() string {
	return fm.ws.String()
}

func (fms *fileMetadataStore) GetSize(layer ChainID) (int64, error) {
	content, err := os.ReadFile(fms.getLayerFilename(layer, "size"))
	if err != nil {
		return 0, err
	}

	size, err := strconv.ParseInt(string(content), 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (fms *fileMetadataStore) GetParent(layer ChainID) (ChainID, error) {
	content, err := os.ReadFile(fms.getLayerFilename(layer, "parent"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	dgst, err := digest.Parse(strings.TrimSpace(string(content)))
	if err != nil {
		return "", err
	}

	return dgst, nil
}

func (fms *fileMetadataStore) GetDiffID(layer ChainID) (DiffID, error) {
	content, err := os.ReadFile(fms.getLayerFilename(layer, "diff"))
	if err != nil {
		return "", err
	}

	dgst, err := digest.Parse(strings.TrimSpace(string(content)))
	if err != nil {
		return "", err
	}

	return dgst, nil
}

func (fms *fileMetadataStore) GetCacheID(layer ChainID) (string, error) {
	contentBytes, err := os.ReadFile(fms.getLayerFilename(layer, "cache-id"))
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))

	if content == "" {
		return "", errors.Errorf("invalid cache id value")
	}

	return content, nil
}

func (fms *fileMetadataStore) GetDescriptor(layer ChainID) (distribution.Descriptor, error) {
	content, err := os.ReadFile(fms.getLayerFilename(layer, "descriptor.json"))
	if err != nil {
		if os.IsNotExist(err) {
			// only return empty descriptor to represent what is stored
			return distribution.Descriptor{}, nil
		}
		return distribution.Descriptor{}, err
	}

	var ref distribution.Descriptor
	err = json.Unmarshal(content, &ref)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	return ref, err
}

func (fms *fileMetadataStore) TarSplitReader(layer ChainID) (io.ReadCloser, error) {
	fz, err := os.Open(fms.getLayerFilename(layer, "tar-split.json.gz"))
	if err != nil {
		return nil, err
	}
	f, err := gzip.NewReader(fz)
	if err != nil {
		fz.Close()
		return nil, err
	}

	return ioutils.NewReadCloserWrapper(f, func() error {
		f.Close()
		return fz.Close()
	}), nil
}

func (fms *fileMetadataStore) SetMountID(mount string, mountID string) error {
	if err := os.MkdirAll(fms.getMountDirectory(mount), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fms.getMountFilename(mount, "mount-id"), []byte(mountID), 0o644)
}

func (fms *fileMetadataStore) SetInitID(mount string, init string) error {
	if err := os.MkdirAll(fms.getMountDirectory(mount), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fms.getMountFilename(mount, "init-id"), []byte(init), 0o644)
}

func (fms *fileMetadataStore) SetMountParent(mount string, parent ChainID) error {
	if err := os.MkdirAll(fms.getMountDirectory(mount), 0o755); err != nil {
		return err
	}
	return os.WriteFile(fms.getMountFilename(mount, "parent"), []byte(parent.String()), 0o644)
}

func (fms *fileMetadataStore) GetMountID(mount string) (string, error) {
	contentBytes, err := os.ReadFile(fms.getMountFilename(mount, "mount-id"))
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))

	if !isValidID(content) {
		return "", errors.New("invalid mount id value")
	}

	return content, nil
}

func (fms *fileMetadataStore) GetInitID(mount string) (string, error) {
	contentBytes, err := os.ReadFile(fms.getMountFilename(mount, "init-id"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	content := strings.TrimSpace(string(contentBytes))

	if !isValidID(content) {
		return "", errors.New("invalid init id value")
	}

	return content, nil
}

func (fms *fileMetadataStore) GetMountParent(mount string) (ChainID, error) {
	content, err := os.ReadFile(fms.getMountFilename(mount, "parent"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	dgst, err := digest.Parse(strings.TrimSpace(string(content)))
	if err != nil {
		return "", err
	}

	return dgst, nil
}

func (fms *fileMetadataStore) getOrphan() ([]roLayer, error) {
	var orphanLayers []roLayer
	for _, algorithm := range supportedAlgorithms {
		fileInfos, err := os.ReadDir(filepath.Join(fms.root, string(algorithm)))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		for _, fi := range fileInfos {
			if !fi.IsDir() || !strings.HasSuffix(fi.Name(), "-removing") {
				continue
			}
			// At this stage, fi.Name value looks like <digest>-<random>-removing
			// Split on '-' to get the digest value.
			nameSplit := strings.Split(fi.Name(), "-")
			dgst := digest.NewDigestFromEncoded(algorithm, nameSplit[0])
			if err := dgst.Validate(); err != nil {
				log.G(context.TODO()).WithError(err).WithField("digest", string(algorithm)+":"+nameSplit[0]).Debug("ignoring invalid digest")
				continue
			}

			chainFile := filepath.Join(fms.root, string(algorithm), fi.Name(), "cache-id")
			contentBytes, err := os.ReadFile(chainFile)
			if err != nil {
				if !os.IsNotExist(err) {
					log.G(context.TODO()).WithError(err).WithField("digest", dgst).Error("failed to read cache ID")
				}
				continue
			}
			cacheID := strings.TrimSpace(string(contentBytes))
			if cacheID == "" {
				log.G(context.TODO()).Error("invalid cache ID")
				continue
			}

			l := &roLayer{
				chainID: dgst,
				cacheID: cacheID,
			}
			orphanLayers = append(orphanLayers, *l)
		}
	}

	return orphanLayers, nil
}

func (fms *fileMetadataStore) List() ([]ChainID, []string, error) {
	var ids []ChainID
	for _, algorithm := range supportedAlgorithms {
		fileInfos, err := os.ReadDir(filepath.Join(fms.root, string(algorithm)))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}

		for _, fi := range fileInfos {
			if fi.IsDir() && fi.Name() != "mounts" {
				dgst := digest.NewDigestFromEncoded(algorithm, fi.Name())
				if err := dgst.Validate(); err != nil {
					log.G(context.TODO()).Debugf("Ignoring invalid digest %s:%s", algorithm, fi.Name())
				} else {
					ids = append(ids, dgst)
				}
			}
		}
	}

	fileInfos, err := os.ReadDir(filepath.Join(fms.root, "mounts"))
	if err != nil {
		if os.IsNotExist(err) {
			return ids, []string{}, nil
		}
		return nil, nil, err
	}

	var mounts []string
	for _, fi := range fileInfos {
		if fi.IsDir() {
			mounts = append(mounts, fi.Name())
		}
	}

	return ids, mounts, nil
}

// Remove layerdb folder if that is marked for removal
func (fms *fileMetadataStore) Remove(layer ChainID, cache string) error {
	dgst := layer
	files, err := os.ReadDir(filepath.Join(fms.root, string(dgst.Algorithm())))
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "-removing") || !strings.HasPrefix(f.Name(), dgst.Encoded()) {
			continue
		}

		// Make sure that we only remove layerdb folder which points to
		// requested cacheID
		dir := filepath.Join(fms.root, string(dgst.Algorithm()), f.Name())
		chainFile := filepath.Join(dir, "cache-id")
		contentBytes, err := os.ReadFile(chainFile)
		if err != nil {
			log.G(context.TODO()).WithError(err).WithField("file", chainFile).Error("cannot get cache ID")
			continue
		}
		cacheID := strings.TrimSpace(string(contentBytes))
		if cacheID != cache {
			continue
		}
		log.G(context.TODO()).Debugf("Removing folder: %s", dir)
		err = os.RemoveAll(dir)
		if err != nil && !os.IsNotExist(err) {
			log.G(context.TODO()).WithError(err).WithField("name", f.Name()).Error("cannot remove layer")
			continue
		}
	}
	return nil
}

func (fms *fileMetadataStore) RemoveMount(mount string) error {
	return os.RemoveAll(fms.getMountDirectory(mount))
}

// isValidID checks if mount/init id is valid. It is similar to
// regexp.MustCompile(`^[a-f0-9]{64}(-init)?$`).MatchString(id).
func isValidID(id string) bool {
	id = strings.TrimSuffix(id, "-init")
	if len(id) != 64 {
		return false
	}
	for _, c := range id {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
