package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/tests"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInMemoryCache_Set_Get_Success(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "test-key"
	value := "test-value"
	duration := 5 * time.Minute

	c.Set(key, value, duration)

	result, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, value, result)
}

func TestInMemoryCache_Get_NotFound(t *testing.T) {
	c := cache.NewInMemoryCache()

	result, found := c.Get("non-existent-key")
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result, got non-nil")
	}
}

func TestInMemoryCache_Set_Get_NoExpiration(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "test-key"
	value := "test-value"

	c.Set(key, value, 0)

	result, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, value, result)
}

func TestInMemoryCache_Expiration(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "test-key"
	value := "test-value"
	duration := 1 * time.Millisecond

	c.Set(key, value, duration)

	result, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, value, result)

	time.Sleep(2 * time.Millisecond)

	result, found = c.Get(key)
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result for expired value, got non-nil")
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "test-key"
	value := "test-value"
	duration := 5 * time.Minute

	c.Set(key, value, duration)

	result, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, value, result)

	c.Delete(key)

	result, found = c.Get(key)
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result after delete, got non-nil")
	}
}

func TestInMemoryCache_Overwrite(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "test-key"
	value1 := "test-value-1"
	value2 := "test-value-2"
	duration := 5 * time.Minute

	c.Set(key, value1, duration)

	result, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, value1, result)

	c.Set(key, value2, duration)

	result, found = c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, value2, result)
}

func TestInMemoryCache_Multiple_Keys(t *testing.T) {
	c := cache.NewInMemoryCache()

	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	duration := 5 * time.Minute

	for key, value := range testData {
		c.Set(key, value, duration)
	}

	for key, expectedValue := range testData {
		result, found := c.Get(key)
		tests.AssertEqual(t, true, found)
		tests.AssertEqual(t, expectedValue, result)
	}

	c.Delete("key2")

	result, found := c.Get("key2")
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result for deleted key2, got non-nil")
	}

	result, found = c.Get("key1")
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, "value1", result)

	result, found = c.Get("key3")
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, "value3", result)
}

func TestInMemoryCache_Complex_Objects(t *testing.T) {
	c := cache.NewInMemoryCache()

	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "password123",
	}
	key := "user:" + user.ID.String()
	duration := 5 * time.Minute

	c.Set(key, user, duration)

	result, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertNotNil(t, result)

	cachedUser, ok := result.(*model.User)
	tests.AssertEqual(t, true, ok)
	tests.AssertEqual(t, user.ID, cachedUser.ID)
	tests.AssertEqual(t, user.Username, cachedUser.Username)
}

func TestInMemoryCache_GetOrSet_CacheHit(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "test-key"
	expectedValue := "cached-value"
	c.Set(key, expectedValue, 5*time.Minute)

	fetcherCalled := false
	fetcher := func() (string, error) {
		fetcherCalled = true
		return "fetcher-value", nil
	}

	result, err := cache.GetOrSet(c, key, 5*time.Minute, fetcher)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, expectedValue, result)
	tests.AssertEqual(t, false, fetcherCalled)
}

func TestInMemoryCache_GetOrSet_CacheMiss(t *testing.T) {
	c := cache.NewInMemoryCache()

	fetcherCalled := false
	expectedValue := "fetcher-value"
	fetcher := func() (string, error) {
		fetcherCalled = true
		return expectedValue, nil
	}

	key := "test-key"

	result, err := cache.GetOrSet(c, key, 5*time.Minute, fetcher)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, expectedValue, result)
	tests.AssertEqual(t, true, fetcherCalled)

	cachedResult, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertEqual(t, expectedValue, cachedResult)
}

func TestInMemoryCache_GetOrSet_FetcherError(t *testing.T) {
	c := cache.NewInMemoryCache()

	fetcher := func() (string, error) {
		return "", tests.ErrorForTesting("fetcher error")
	}

	key := "test-key"

	result, err := cache.GetOrSet(c, key, 5*time.Minute, fetcher)
	tests.AssertError(t, err, "")
	tests.AssertEqual(t, "", result)

	cachedResult, found := c.Get(key)
	tests.AssertEqual(t, false, found)
	if cachedResult != nil {
		t.Fatal("Expected nil cachedResult, got non-nil")
	}
}

func TestInMemoryCache_TypeSafety(t *testing.T) {
	c := cache.NewInMemoryCache()

	userFetcher := func() (*model.User, error) {
		return &model.User{
			ID:       uuid.New(),
			Username: "testuser",
		}, nil
	}

	key := "user-key"

	user, err := cache.GetOrSet(c, key, 5*time.Minute, userFetcher)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, user)
	tests.AssertEqual(t, "testuser", user.Username)

	cachedResult, found := c.Get(key)
	tests.AssertEqual(t, true, found)
	cachedUser, ok := cachedResult.(*model.User)
	tests.AssertEqual(t, true, ok)
	tests.AssertEqual(t, user.ID, cachedUser.ID)
}

func TestInMemoryCache_Concurrent_Access(t *testing.T) {
	c := cache.NewInMemoryCache()

	key := "concurrent-key"
	value := "concurrent-value"
	duration := 5 * time.Minute

	done := make(chan bool, 3)

	go func() {
		c.Set(key, value, duration)
		done <- true
	}()

	go func() {
		time.Sleep(1 * time.Millisecond)
		result, found := c.Get(key)
		if found {
			tests.AssertEqual(t, value, result)
		}
		done <- true
	}()

	go func() {
		time.Sleep(2 * time.Millisecond)
		c.Delete(key)
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}

	result, found := c.Get(key)
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result, got non-nil")
	}
}

func TestServerStatusCache_GetStatus_NeedsRefresh(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   1 * time.Second,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerStatusCache(config)

	serviceName := "test-service"

	status, needsRefresh := cache.GetStatus(serviceName)
	tests.AssertEqual(t, model.StatusUnknown, status)
	tests.AssertEqual(t, true, needsRefresh)
}

func TestServerStatusCache_UpdateStatus_GetStatus(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   1 * time.Second,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerStatusCache(config)

	serviceName := "test-service"
	expectedStatus := model.StatusRunning

	cache.UpdateStatus(serviceName, expectedStatus)

	status, needsRefresh := cache.GetStatus(serviceName)
	tests.AssertEqual(t, expectedStatus, status)
	tests.AssertEqual(t, false, needsRefresh)
}

func TestServerStatusCache_Throttling(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   100 * time.Millisecond,
		DefaultStatus:  model.StatusStopped,
	}
	cache := model.NewServerStatusCache(config)

	serviceName := "test-service"

	cache.UpdateStatus(serviceName, model.StatusRunning)

	status, needsRefresh := cache.GetStatus(serviceName)
	tests.AssertEqual(t, model.StatusRunning, status)
	tests.AssertEqual(t, false, needsRefresh)

	status, needsRefresh = cache.GetStatus(serviceName)
	tests.AssertEqual(t, model.StatusRunning, status)
	tests.AssertEqual(t, false, needsRefresh)

	time.Sleep(150 * time.Millisecond)

	_, _ = cache.GetStatus(serviceName)

}

func TestServerStatusCache_Expiration(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 50 * time.Millisecond,
		ThrottleTime:   10 * time.Millisecond,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerStatusCache(config)

	serviceName := "test-service"
	cache.UpdateStatus(serviceName, model.StatusRunning)

	status, needsRefresh := cache.GetStatus(serviceName)
	tests.AssertEqual(t, model.StatusRunning, status)
	tests.AssertEqual(t, false, needsRefresh)

	time.Sleep(60 * time.Millisecond)

	status, needsRefresh = cache.GetStatus(serviceName)
	tests.AssertEqual(t, true, needsRefresh)
}

func TestServerStatusCache_InvalidateStatus(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   1 * time.Second,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerStatusCache(config)

	serviceName := "test-service"

	cache.UpdateStatus(serviceName, model.StatusRunning)

	status, needsRefresh := cache.GetStatus(serviceName)
	tests.AssertEqual(t, model.StatusRunning, status)
	tests.AssertEqual(t, false, needsRefresh)

	cache.InvalidateStatus(serviceName)

	status, needsRefresh = cache.GetStatus(serviceName)
	tests.AssertEqual(t, model.StatusUnknown, status)
	tests.AssertEqual(t, true, needsRefresh)
}

func TestServerStatusCache_Clear(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   1 * time.Second,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerStatusCache(config)

	services := []string{"service1", "service2", "service3"}
	for _, service := range services {
		cache.UpdateStatus(service, model.StatusRunning)
	}

	for _, service := range services {
		status, needsRefresh := cache.GetStatus(service)
		tests.AssertEqual(t, model.StatusRunning, status)
		tests.AssertEqual(t, false, needsRefresh)
	}

	cache.Clear()

	for _, service := range services {
		status, needsRefresh := cache.GetStatus(service)
		tests.AssertEqual(t, model.StatusUnknown, status)
		tests.AssertEqual(t, true, needsRefresh)
	}
}

func TestLookupCache_SetGetClear(t *testing.T) {
	cache := model.NewLookupCache()

	key := "lookup-key"
	value := map[string]string{"test": "data"}

	cache.Set(key, value)

	result, found := cache.Get(key)
	tests.AssertEqual(t, true, found)
	tests.AssertNotNil(t, result)

	resultMap, ok := result.(map[string]string)
	tests.AssertEqual(t, true, ok)
	tests.AssertEqual(t, "data", resultMap["test"])

	cache.Clear()

	result, found = cache.Get(key)
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result, got non-nil")
	}
}

func TestServerConfigCache_Configuration(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   1 * time.Second,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerConfigCache(config)

	serverID := uuid.New().String()
	configuration := model.Configuration{
		UdpPort:         model.IntString(9231),
		TcpPort:         model.IntString(9232),
		MaxConnections:  model.IntString(30),
		LanDiscovery:    model.IntString(1),
		RegisterToLobby: model.IntString(1),
		ConfigVersion:   model.IntString(1),
	}

	result, found := cache.GetConfiguration(serverID)
	tests.AssertEqual(t, false, found)
	if result != nil {
		t.Fatal("Expected nil result, got non-nil")
	}

	cache.UpdateConfiguration(serverID, configuration)

	result, found = cache.GetConfiguration(serverID)
	tests.AssertEqual(t, true, found)
	tests.AssertNotNil(t, result)
	tests.AssertEqual(t, configuration.UdpPort, result.UdpPort)
	tests.AssertEqual(t, configuration.TcpPort, result.TcpPort)
}

func TestServerConfigCache_InvalidateServerCache(t *testing.T) {
	config := model.CacheConfig{
		ExpirationTime: 5 * time.Minute,
		ThrottleTime:   1 * time.Second,
		DefaultStatus:  model.StatusUnknown,
	}
	cache := model.NewServerConfigCache(config)

	serverID := uuid.New().String()
	configuration := model.Configuration{UdpPort: model.IntString(9231)}
	assistRules := model.AssistRules{StabilityControlLevelMax: model.IntString(0)}

	cache.UpdateConfiguration(serverID, configuration)
	cache.UpdateAssistRules(serverID, assistRules)

	configResult, found := cache.GetConfiguration(serverID)
	tests.AssertEqual(t, true, found)
	tests.AssertNotNil(t, configResult)

	assistResult, found := cache.GetAssistRules(serverID)
	tests.AssertEqual(t, true, found)
	tests.AssertNotNil(t, assistResult)

	cache.InvalidateServerCache(serverID)

	configResult, found = cache.GetConfiguration(serverID)
	tests.AssertEqual(t, false, found)
	if configResult != nil {
		t.Fatal("Expected nil configResult, got non-nil")
	}

	assistResult, found = cache.GetAssistRules(serverID)
	tests.AssertEqual(t, false, found)
	if assistResult != nil {
		t.Fatal("Expected nil assistResult, got non-nil")
	}
}
