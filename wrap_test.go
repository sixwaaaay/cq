/*
 * Copyright (c) 2023 sixwaaaay.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cq

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ErrNotFound = errors.New("not found")

func TestWrapRepoFindOneReturnsExpectedResult(t *testing.T) {
	ctx := context.Background()
	expectedUser := User{ID: 1, Name: "John Doe", Email: "john.doe@example.com"}

	repo := NewWrapRepo[User](
		func(ctx context.Context, id int64) (*User, error) {
			return &expectedUser, nil
		},
		nil,
	)

	result, err := repo.FindOne(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, *result)
}

func TestWrapRepoFindOneReturnsErrorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	repo := NewWrapRepo[User](
		func(ctx context.Context, id int64) (*User, error) {
			return nil, ErrNotFound
		},
		nil,
	)

	_, err := repo.FindOne(ctx, 1)

	assert.ErrorIs(t, err, ErrNotFound)
}

func TestWrapRepoFindManyReturnsExpectedResults(t *testing.T) {
	ctx := context.Background()
	expectedUsers := []User{
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	}

	repo := NewWrapRepo[User](
		nil,
		func(ctx context.Context, ids []int64) ([]User, error) {
			return expectedUsers, nil
		},
	)

	results, err := repo.FindMany(ctx, []int64{1, 2})

	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, results)
}

func TestWrapRepoFindManyReturnsErrorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	repo := NewWrapRepo[User](
		nil,
		func(ctx context.Context, ids []int64) ([]User, error) {
			return nil, ErrNotFound
		},
	)

	_, err := repo.FindMany(ctx, []int64{1, 2})

	assert.ErrorIs(t, err, ErrNotFound)
}
