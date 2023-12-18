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

import "context"

type wrapRepo[T any] struct {
	findOne  func(ctx context.Context, id int64) (*T, error)
	findMany func(ctx context.Context, ids []int64) ([]T, error)
}

func NewWrapRepo[T any](findOne func(ctx context.Context, id int64) (*T, error), findMany func(ctx context.Context, ids []int64) ([]T, error)) wrapRepo[T] {
	return wrapRepo[T]{
		findOne:  findOne,
		findMany: findMany,
	}
}

func (r wrapRepo[T]) FindOne(ctx context.Context, id int64) (*T, error) {
	return r.findOne(ctx, id)
}

func (r wrapRepo[T]) FindMany(ctx context.Context, ids []int64) ([]T, error) {
	return r.findMany(ctx, ids)
}
