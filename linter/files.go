package main

type Language string

const (
    Python Language = "python"
    Java = "java"
)

type SourceFile struct {
    Content string `json:"content"`
}
