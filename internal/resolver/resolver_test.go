package resolver

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/torqio/protodep/internal/auth"
)

func TestSync(t *testing.T) {
	homeDir, err := homedir.Dir()
	require.NoError(t, err)

	dotProtoDir := filepath.Join(homeDir, "protodep_ut")
	err = os.RemoveAll(dotProtoDir)
	require.NoError(t, err)

	pwd, err := os.Getwd()
	fmt.Println(pwd)

	require.NoError(t, err)

	outputRootDir := os.TempDir()

	conf := Config{
		HomeDir:   dotProtoDir,
		TargetDir: pwd,
		OutputDir: outputRootDir,
	}

	c := gomock.NewController(t)
	defer c.Finish()

	httpsAuthProviderMock := auth.NewMockAuthProvider(c)
	httpsAuthProviderMock.EXPECT().AuthMethod().Return(nil, nil).AnyTimes()
	httpsAuthProviderMock.EXPECT().GetRepositoryURL("github.com/protocolbuffers/protobuf").Return("https://github.com/protocolbuffers/protobuf.git")
	httpsAuthProviderMock.EXPECT().GetRepositoryURL("github.com/protodep/catalog").Return("https://github.com/protodep/catalog.git")

	sshAuthProviderMock := auth.NewMockAuthProvider(c)
	sshAuthProviderMock.EXPECT().AuthMethod().Return(nil, nil).AnyTimes()
	sshAuthProviderMock.EXPECT().GetRepositoryURL("github.com/opensaasstudio/plasma").Return("https://github.com/opensaasstudio/plasma.git")

	target, err := New(&conf, httpsAuthProviderMock, sshAuthProviderMock)
	require.NoError(t, err)

	// clone
	err = target.Resolve(false)
	require.NoError(t, err)

	if !isFileExist(filepath.Join(outputRootDir, "proto/stream.proto")) {
		t.Error("not found file [proto/stream.proto]")
	}
	if !isFileExist(filepath.Join(outputRootDir, "proto/google/protobuf/empty.proto")) {
		t.Error("not found file [proto/google/protobuf/empty.proto]")
	}

	// check ignore worked
	// hasPrefix test - backward compatibility
	if isFileExist(filepath.Join(outputRootDir, "proto/google/protobuf/test_messages_proto3.proto")) {
		t.Error("found file [proto/google/protobuf/test_messages_proto3.proto]")
	}

	// glob test 1
	if isFileExist(filepath.Join(outputRootDir, "proto/google/protobuf/test_messages_proto2.proto")) {
		t.Error("found file [proto/google/protobuf/test_messages_proto2.proto]")
	}

	// glob test 2
	if isFileExist(filepath.Join(outputRootDir, "proto/google/protobuf/test_messages_proto2.proto")) {
		t.Error("found file [proto/google/protobuf/test_messages_proto2.proto]")
	}

	// glob test 3
	if isFileExist(filepath.Join(outputRootDir, "proto/google/protobuf/util/internal/testdata/")) {
		t.Error("found file [proto/google/protobuf/util/internal/testdata/]")
	}

	// check include worked
	// glob test 1
	if !isFileExist(filepath.Join(outputRootDir, "proto/protodep/hierarchy/service.proto")) {
		t.Error("not found file [proto/protodep/hierarchy/service.proto]")
	}

	// glob test 2
	if !isFileExist(filepath.Join(outputRootDir, "proto/protodep/hierarchy/fuga/fuga.proto")) {
		t.Error("not found file [proto/protodep/hierarchy/fuga/fuga.proto]")
	}

	// fetch
	err = target.Resolve(false)
	require.NoError(t, err)
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TestWriteFileWithDirectory(t *testing.T) {
	destDir := os.TempDir()
	testDir := filepath.Join(destDir, "hoge")
	testFile := filepath.Join(testDir, "fuga.txt")

	err := writeFileWithDirectory(testFile, []byte("test"), 0o644)
	require.NoError(t, err)

	stat, err := os.Stat(testFile)
	require.NoError(t, err)
	require.True(t, !stat.IsDir())

	data, err := os.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, string(data), "test")
}

func TestIsAvailableSSH(t *testing.T) {
	f, err := os.CreateTemp("", "id_rsa")
	require.NoError(t, err)

	found, err := isAvailableSSH(f.Name())
	require.NoError(t, err)
	require.True(t, found)

	notFound, err := isAvailableSSH(fmt.Sprintf("/tmp/IsAvailableSSH_%d", time.Now().UnixNano()))
	require.NoError(t, err)
	require.False(t, notFound)
}
