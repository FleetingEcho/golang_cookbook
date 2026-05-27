# 前缀树 — Trie

```go
type TrieNode struct {
    children [26]*TrieNode // 小写字母，可用 map 支持任意字符
    isEnd    bool
}

type Trie struct {
    root *TrieNode
}

func NewTrie() *Trie {
    return &Trie{root: &TrieNode{}}
}

func (t *Trie) Insert(word string) {
    node := t.root
    for _, ch := range word {
        idx := ch - 'a'
        if node.children[idx] == nil {
            node.children[idx] = &TrieNode{}
        }
        node = node.children[idx]
    }
    node.isEnd = true
}

func (t *Trie) Search(word string) bool {
    node := t.find(word)
    return node != nil && node.isEnd
}

func (t *Trie) StartsWith(prefix string) bool {
    return t.find(prefix) != nil
}

func (t *Trie) find(s string) *TrieNode {
    node := t.root
    for _, ch := range s {
        idx := ch - 'a'
        if node.children[idx] == nil {
            return nil
        }
        node = node.children[idx]
    }
    return node
}

// 按前缀搜索所有单词
func (t *Trie) SearchByPrefix(prefix string) []string {
    node := t.find(prefix)
    if node == nil { return nil }

    var result []string
    var dfs func(*TrieNode, string)
    dfs = func(n *TrieNode, s string) {
        if n.isEnd {
            result = append(result, s)
        }
        for i, child := range n.children {
            if child != nil {
                dfs(child, s+string(rune('a'+i)))
            }
        }
    }
    dfs(node, prefix)
    return result
}

// 泛型 Trie（支持任意可哈希键）
type GenericTrieNode[T comparable] struct {
    children map[T]*GenericTrieNode[T]
    isEnd    bool
}

type GenericTrie[T comparable] struct {
    root *GenericTrieNode[T]
}

func NewGenericTrie[T comparable]() *GenericTrie[T] {
    return &GenericTrie[T]{root: &GenericTrieNode[T]{children: make(map[T]*GenericTrieNode[T])}}
}

func (t *GenericTrie[T]) Insert(keys []T) {
    node := t.root
    for _, k := range keys {
        if node.children[k] == nil {
            node.children[k] = &GenericTrieNode[T]{children: make(map[T]*GenericTrieNode[T])}
        }
        node = node.children[k]
    }
    node.isEnd = true
}

func (t *GenericTrie[T]) Search(keys []T) bool {
    node := t.root
    for _, k := range keys {
        if node.children[k] == nil { return false }
        node = node.children[k]
    }
    return node.isEnd
}
```

## 完整对照表

| 操作 | TS | Go |
|------|-----|-----|
| 节点 | `class TrieNode { children: Map }` | `type TrieNode struct { children [26]*TrieNode }` |
| 插入 | `insert(word)` | `Insert(word)` |
| 搜索 | `search(word)` | `Search(word)` |
| 前缀 | `startsWith(prefix)` | `StartsWith(prefix)` |
| 字母 | 任意 | 定长数组（26字母）或 map |
| 泛型 | `Trie<T>` | `GenericTrie[T comparable]` |
