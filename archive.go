package main

import (
	"encoding/json"
	"io"
	"os"
)

type TrainID string

type TrainArchive struct {
	hash map[string]trainArchiveValue
}
type trainArchiveValue struct {
	MessageID int
	TrainHash string
}

func LoadTrainArchive(r io.Reader) (*TrainArchive, error) {
	var ta TrainArchive
	err := json.NewDecoder(r).Decode(&ta)
	return &ta, err
}

type TrainArchiveCompare int

const (
	TrainNotSaved TrainArchiveCompare = iota
	TrainChanged
	TrainSaved
)

func LoadTrainArchiveFromFile(file string) (*TrainArchive, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		os.WriteFile(file, []byte("{}"), 0655)
	}
	fl, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return LoadTrainArchive(fl)
}

func (t *TrainArchive) SaveAsFile(file string) error {
	fl, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0655)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(t, "", "\t")
	if err != nil {
		return err
	}

	_, err = fl.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (t *TrainArchive) Add(train Train, msgID int) {
	if t.hash == nil {
		t.hash = make(map[string]trainArchiveValue)
	}

	t.hash[train.UniqueID()] = trainArchiveValue{
		MessageID: msgID,
		TrainHash: train.Hash(),
	}
}

func (t *TrainArchive) IsSaved(train Train) bool {
	_, found := t.hash[train.UniqueID()]
	return found
}
func (t *TrainArchive) GetID(train Train) int {
	return t.hash[train.UniqueID()].MessageID
}

func (t *TrainArchive) Compare(new Train) TrainArchiveCompare {
	old, found := t.hash[new.UniqueID()]
	if !found {
		return TrainNotSaved
	}
	if new.Hash() != old.TrainHash {
		return TrainChanged
	}

	return TrainSaved
}

func (t *TrainArchive) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.hash)
}

func (t *TrainArchive) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.hash)
}
