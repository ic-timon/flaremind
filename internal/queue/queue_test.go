package queue

import (
	"testing"
)

func TestQueue(t *testing.T) {
	q := NewQueue()

	// 测试 Add
	if !q.Add("https://example.com/page1") {
		t.Error("Expected to add URL to queue")
	}

	if q.Add("https://example.com/page1") {
		t.Error("Expected not to add duplicate URL")
	}

	// 测试 Size
	if q.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", q.Size())
	}

	// 测试 Pop
	url, ok := q.Pop()
	if !ok {
		t.Error("Expected to pop URL from queue")
	}
	if url != "https://example.com/page1" {
		t.Errorf("Expected https://example.com/page1, got %s", url)
	}

	// 测试空队列
	_, ok = q.Pop()
	if ok {
		t.Error("Expected not to pop from empty queue")
	}

	// 测试 MarkVisited 和 IsVisited
	q.Add("https://example.com/page2")
	q.MarkVisited("https://example.com/page2")
	if !q.IsVisited("https://example.com/page2") {
		t.Error("Expected URL to be marked as visited")
	}

	if q.VisitedCount() != 1 {
		t.Errorf("Expected 1 visited URL, got %d", q.VisitedCount())
	}
}


