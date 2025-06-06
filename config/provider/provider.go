// Copyright 2021 Meshery Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file needs a better name.

// Package provider provides config provider implementations that can be used in the adapters, as well as the Options type containing options for various aspects of an adapter.
package provider

const (
	// Provider keys
	ViperKey = "viper"
	InMemKey = "in-mem"
)

// Type Options contains config options for various aspects of an adapter.
type Options struct {
	FilePath string
	FileType string
	FileName string
}
