package watch

import "testing"

func TestEventBroadcasterRoutingAndSlowSubscriber(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	a, aDone, unsubscribeA := broadcaster.Subscribe("cluster-a")
	defer unsubscribeA()
	b, bDone, unsubscribeB := broadcaster.Subscribe("cluster-b")
	defer unsubscribeB()

	assertEvent := func(events <-chan Event, want string) {
		t.Helper()
		select {
		case event := <-events:
			if event.Type != want {
				t.Fatalf("received %q, want %q", event.Type, want)
			}
		default:
			t.Fatalf("missing event %q", want)
		}
	}

	broadcaster.Publish("cluster-a", Event{Type: "clusterCR"})
	assertEvent(a, "clusterCR")
	select {
	case event := <-b:
		t.Fatalf("cluster-b received %q for cluster-a", event.Type)
	default:
	}
	broadcaster.Publish("", Event{Type: "management"})
	assertEvent(a, "management")
	assertEvent(b, "management")

	for range cap(a) + 1 {
		broadcaster.Publish("", Event{Type: "management"})
		assertEvent(b, "management")
	}
	select {
	case <-aDone:
	default:
		t.Fatal("slow subscriber was not disconnected")
	}
	select {
	case <-bDone:
		t.Fatal("healthy subscriber was disconnected")
	default:
	}
	broadcaster.Publish("", Event{Type: "management"})
	assertEvent(b, "management")
}
