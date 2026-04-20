package eventbus

import (
	"reflect"
	"sync"
	"testing"
)

func TestGetReturnsSameBusForSameType(t *testing.T) {
	resetManagerForTest(t)

	first := Get[UserCreated]()
	second := Get[UserCreated]()

	if first != second {
		t.Fatalf("expected same bus instance for same event type")
	}
}

func TestGetReturnsDifferentBusForDifferentTypes(t *testing.T) {
	resetManagerForTest(t)

	userBus := Get[UserCreated]()
	orderBus := Get[OrderCreated]()

	if userBus == any(orderBus) {
		t.Fatalf("expected different bus instances for different event types")
	}
}

func TestGetCreatesOnlyOneBusUnderConcurrency(t *testing.T) {
	resetManagerForTest(t)

	const workers = 32

	results := make(chan *Bus[UserCreated], workers)
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- Get[UserCreated]()
		}()
	}

	wg.Wait()
	close(results)

	var first *Bus[UserCreated]
	for bus := range results {
		if first == nil {
			first = bus
			continue
		}
		if bus != first {
			t.Fatalf("expected same bus instance for concurrent access")
		}
	}
}

func resetManagerForTest(t *testing.T) {
	t.Helper()

	mu.Lock()
	buss = make(map[reflect.Type]any)
	mu.Unlock()
}
