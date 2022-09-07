package main

import (
	"encoding/json"
	"io"
	"os"
)

type Hashable interface{ Hash() string }
type HashSet[T Hashable] struct{ hash map[string]struct{} }

func LoadHashSet[T Hashable](r io.Reader) (*HashSet[T], error) {
	hashSet := &HashSet[T]{}

	err := json.NewDecoder(r).Decode(hashSet)

	return hashSet, err
}

func LoadHashSetFromFile[T Hashable](file string) (*HashSet[T], error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.WriteFile(file, []byte("{}"), 0655)
	}
	fl, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return LoadHashSet[T](fl)
}

func (h HashSet[T]) SaveAsFile(file string) error {
	fl, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0655)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(h)
	if err != nil {
		return err
	}

	_, err = fl.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (h *HashSet[T]) Add(what T) {
	if h.hash == nil {
		h.hash = make(map[string]struct{})
	}

	h.hash[what.Hash()] = struct{}{}
}

func (h HashSet[T]) IsSaved(what T) bool {
	_, found := h.hash[what.Hash()]
	return found
}

func (h HashSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.hash)
}

func (h *HashSet[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &h.hash)
}
