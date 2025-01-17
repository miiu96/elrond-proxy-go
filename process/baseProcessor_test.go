package process_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-proxy-go/data"
	"github.com/ElrondNetwork/elrond-proxy-go/process"
	"github.com/ElrondNetwork/elrond-proxy-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Nonce int
	Name  string
}

func createTestHttpServer(
	matchingPath string,
	response []byte,
) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			if req.URL.String() == matchingPath {
				_, _ = rw.Write(response)
			}
		}

		if req.Method == "POST" {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(req.Body)
			_, _ = rw.Write(buf.Bytes())
		}
	}))
}

func TestNewBaseProcessor_WithInvalidRequestTimeoutShouldErr(t *testing.T) {
	t.Parallel()

	bp, err := process.NewBaseProcessor(
		-5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	assert.Nil(t, bp)
	assert.Equal(t, process.ErrInvalidRequestTimeout, err)
}

func TestNewBaseProcessor_WithNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	bp, err := process.NewBaseProcessor(
		5,
		nil,
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	assert.Nil(t, bp)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewBaseProcessor_WithNilObserversProviderShouldErr(t *testing.T) {
	t.Parallel()

	bp, err := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		nil,
		&mock.PubKeyConverterMock{},
	)

	assert.Nil(t, bp)
	assert.True(t, errors.Is(err, process.ErrNilNodesProvider))
}

func TestNewBaseProcessor_WithNilFullHistoryNodesProviderShouldErr(t *testing.T) {
	t.Parallel()

	bp, err := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		nil,
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	assert.Nil(t, bp)
	assert.True(t, errors.Is(err, process.ErrNilNodesProvider))
}

func TestNewBaseProcessor_WithOkValuesShouldWork(t *testing.T) {
	t.Parallel()

	bp, err := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	assert.NotNil(t, bp)
	assert.Nil(t, err)
}

//------- GetObservers

func TestBaseProcessor_GetObserversEmptyListShouldWork(t *testing.T) {
	t.Parallel()

	observersSlice := []*data.NodeData{{Address: "addr1"}}
	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{
			GetNodesByShardIdCalled: func(_ uint32) ([]*data.NodeData, error) {
				return observersSlice, nil
			},
		},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)
	observers, err := bp.GetObservers(0)

	assert.Nil(t, err)
	assert.Equal(t, observersSlice, observers)
}

//------- ComputeShardId

func TestBaseProcessor_ComputeShardId(t *testing.T) {
	t.Parallel()

	observersList := []*data.NodeData{
		{
			Address: "address1",
			ShardId: 0,
		},
		{
			Address: "address2",
			ShardId: 1,
		},
	}

	msc, _ := sharding.NewMultiShardCoordinator(3, 0)
	bp, _ := process.NewBaseProcessor(
		5,
		msc,
		&mock.ObserversProviderStub{
			GetNodesByShardIdCalled: func(_ uint32) ([]*data.NodeData, error) {
				return observersList, nil
			},
		},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	//there are 2 shards, compute ID should correctly process
	addressInShard0 := []byte{0}
	shardID, err := bp.ComputeShardId(addressInShard0)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), shardID)

	addressInShard1 := []byte{1}
	shardID, err = bp.ComputeShardId(addressInShard1)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), shardID)
}

//------- Calls

func TestBaseProcessor_CallGetRestEndPoint(t *testing.T) {
	ts := &testStruct{
		Nonce: 10000,
		Name:  "a test struct to be send and received",
	}
	response, _ := json.Marshal(ts)

	server := createTestHttpServer("/some/path", response)
	fmt.Printf("Server: %s\n", server.URL)
	defer server.Close()

	tsRecovered := &testStruct{}
	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)
	_, err := bp.CallGetRestEndPoint(server.URL, "/some/path", tsRecovered)

	assert.Nil(t, err)
	assert.Equal(t, ts, tsRecovered)
}

func TestBaseProcessor_CallGetRestEndPointShouldTimeout(t *testing.T) {
	ts := &testStruct{
		Nonce: 10000,
		Name:  "a test struct to be send and received",
	}
	response, _ := json.Marshal(ts)

	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		_, _ = rw.Write(response)
	}))
	fmt.Printf("Server: %s\n", testServer.URL)
	defer testServer.Close()

	tsRecovered := &testStruct{}
	bp, _ := process.NewBaseProcessor(
		1,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)
	_, err := bp.CallGetRestEndPoint(testServer.URL, "/some/path", tsRecovered)

	assert.NotEqual(t, ts.Name, tsRecovered.Name)
	assert.NotNil(t, err)
}

func TestBaseProcessor_CallPostRestEndPoint(t *testing.T) {
	ts := &testStruct{
		Nonce: 10000,
		Name:  "a test struct to be send",
	}
	tsRecv := &testStruct{}

	server := createTestHttpServer("/some/path", nil)
	fmt.Printf("Server: %s\n", server.URL)
	defer server.Close()

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)
	rc, err := bp.CallPostRestEndPoint(server.URL, "/some/path", ts, tsRecv)

	assert.Nil(t, err)
	assert.Equal(t, ts, tsRecv)
	assert.Equal(t, http.StatusOK, rc)
}

func TestBaseProcessor_CallPostRestEndPointShouldTimeout(t *testing.T) {
	ts := &testStruct{
		Nonce: 10000,
		Name:  "a test struct to be send",
	}
	tsRecv := &testStruct{}

	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		tsBytes, _ := json.Marshal(ts)
		_, _ = rw.Write(tsBytes)
	}))

	fmt.Printf("Server: %s\n", testServer.URL)
	defer testServer.Close()

	bp, _ := process.NewBaseProcessor(
		1,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)
	rc, err := bp.CallPostRestEndPoint(testServer.URL, "/some/path", ts, tsRecv)

	assert.NotEqual(t, tsRecv.Name, ts.Name)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusRequestTimeout, rc)
}

func TestBaseProcessor_GetAllObserversWithOkValuesShouldPass(t *testing.T) {
	t.Parallel()

	statusResponse := data.StatusResponse{
		Message: "",
		Error:   "",
		Running: true,
	}

	statusResponseBytes, err := json.Marshal(statusResponse)
	assert.Nil(t, err)

	server := createTestHttpServer("/node/status", statusResponseBytes)
	fmt.Printf("Server: %s\n", server.URL)
	defer server.Close()

	var observersList []*data.NodeData
	observersList = append(observersList, &data.NodeData{
		ShardId: 0,
		Address: server.URL,
	})

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{},
		&mock.ObserversProviderStub{
			GetAllNodesCalled: func() ([]*data.NodeData, error) {
				return observersList, nil
			},
		},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	assert.Nil(t, err)

	observers, _ := bp.GetAllObservers()
	assert.Nil(t, err)
	assert.Equal(t, server.URL, observers[0].Address)
}

func TestBaseProcessor_GetObserversOnePerShardShouldWork(t *testing.T) {
	t.Parallel()

	expectedResult := []string{
		"shard 0 - id 0",
		"shard 1 - id 0",
		"shard meta - id 0",
	}

	observersListShard0 := []*data.NodeData{
		{Address: "shard 0 - id 0"},
		{Address: "shard 0 - id 1"},
	}
	observersListShard1 := []*data.NodeData{
		{Address: "shard 1 - id 0"},
		{Address: "shard 1 - id 1"},
	}
	observersListShardMeta := []*data.NodeData{
		{Address: "shard meta - id 0"},
		{Address: "shard meta - id 1"},
	}

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{NumShards: 2},
		&mock.ObserversProviderStub{
			GetNodesByShardIdCalled: func(shardId uint32) ([]*data.NodeData, error) {
				switch shardId {
				case 0:
					return observersListShard0, nil
				case 1:
					return observersListShard1, nil
				case core.MetachainShardId:
					return observersListShardMeta, nil
				}

				return nil, nil
			},
		},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	observers, err := bp.GetObserversOnePerShard()
	assert.NoError(t, err)

	for i := 0; i < len(observers); i++ {
		assert.Equal(t, expectedResult[i], observers[i].Address)
	}
	assert.Equal(t, len(expectedResult), len(observers))
}

func TestBaseProcessor_GetObserversOnePerShardOneShardHasNoObserverShouldWork(t *testing.T) {
	t.Parallel()

	expectedResult := []string{
		"shard 0 - id 0",
		"shard meta - id 0",
	}

	observersListShard0 := []*data.NodeData{
		{Address: "shard 0 - id 0"},
		{Address: "shard 0 - id 1"},
	}
	var observersListShard1 []*data.NodeData
	observersListShardMeta := []*data.NodeData{
		{Address: "shard meta - id 0"},
		{Address: "shard meta - id 1"},
	}

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{NumShards: 2},
		&mock.ObserversProviderStub{
			GetNodesByShardIdCalled: func(shardId uint32) ([]*data.NodeData, error) {
				switch shardId {
				case 0:
					return observersListShard0, nil
				case 1:
					return observersListShard1, nil
				case core.MetachainShardId:
					return observersListShardMeta, nil
				}

				return nil, nil
			},
		},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	observers, err := bp.GetObserversOnePerShard()
	assert.NoError(t, err)

	for i := 0; i < len(observers); i++ {
		assert.Equal(t, expectedResult[i], observers[i].Address)
	}
	assert.Equal(t, len(expectedResult), len(observers))
}

func TestBaseProcessor_GetObserversOnePerShardMetachainHasNoObserverShouldWork(t *testing.T) {
	t.Parallel()

	expectedResult := []string{
		"shard 0 - id 0",
		"shard 1 - id 0",
	}

	observersListShard0 := []*data.NodeData{
		{Address: "shard 0 - id 0"},
		{Address: "shard 0 - id 1"},
	}
	observersListShard1 := []*data.NodeData{
		{Address: "shard 1 - id 0"},
		{Address: "shard 1 - id 0"},
	}
	var observersListShardMeta []*data.NodeData

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{NumShards: 2},
		&mock.ObserversProviderStub{
			GetNodesByShardIdCalled: func(shardId uint32) ([]*data.NodeData, error) {
				switch shardId {
				case 0:
					return observersListShard0, nil
				case 1:
					return observersListShard1, nil
				case core.MetachainShardId:
					return observersListShardMeta, nil
				}

				return nil, nil
			},
		},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	observers, err := bp.GetObserversOnePerShard()
	assert.NoError(t, err)

	for i := 0; i < len(observers); i++ {
		assert.Equal(t, expectedResult[i], observers[i].Address)
	}
	assert.Equal(t, len(expectedResult), len(observers))
}

func TestBaseProcessor_GetFullHistoryNodesOnePerShardShouldWork(t *testing.T) {
	t.Parallel()

	expectedResult := []string{
		"shard 0 - id 0",
		"shard 1 - id 0",
		"shard meta - id 0",
	}

	observersListShard0 := []*data.NodeData{
		{Address: "shard 0 - id 0"},
		{Address: "shard 0 - id 1"},
	}
	observersListShard1 := []*data.NodeData{
		{Address: "shard 1 - id 0"},
		{Address: "shard 1 - id 1"},
	}
	observersListShardMeta := []*data.NodeData{
		{Address: "shard meta - id 0"},
		{Address: "shard meta - id 1"},
	}

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{NumShards: 2},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{
			GetNodesByShardIdCalled: func(shardId uint32) ([]*data.NodeData, error) {
				switch shardId {
				case 0:
					return observersListShard0, nil
				case 1:
					return observersListShard1, nil
				case core.MetachainShardId:
					return observersListShardMeta, nil
				}

				return nil, nil
			},
		},
		&mock.PubKeyConverterMock{},
	)

	observers, err := bp.GetFullHistoryNodesOnePerShard()
	assert.NoError(t, err)

	for i := 0; i < len(observers); i++ {
		assert.Equal(t, expectedResult[i], observers[i].Address)
	}
	assert.Equal(t, len(expectedResult), len(observers))
}

func TestBaseProcessor_GetShardIDs(t *testing.T) {
	t.Parallel()

	bp, _ := process.NewBaseProcessor(
		5,
		&mock.ShardCoordinatorMock{NumShards: 3},
		&mock.ObserversProviderStub{},
		&mock.ObserversProviderStub{},
		&mock.PubKeyConverterMock{},
	)

	expected := []uint32{0, 1, 2, core.MetachainShardId}
	require.Equal(t, expected, bp.GetShardIDs())
}
