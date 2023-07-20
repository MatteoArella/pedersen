// Copyright (c) Pedersen authors.
// 
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

#include "goopenssl.h"

#include <stdio.h>
#include <sys/types.h>

#if defined(_WIN32) || defined(_WIN64)

#include <windows.h>

#define MUTEX_TYPE HANDLE
#define MUTEX_SETUP(x) (x = CreateMutex(NULL, FALSE, NULL))
#define MUTEX_CLEANUP(x) CloseHandle(x)
#define MUTEX_LOCK(x) WaitForSingleObject(x, INFINITE)
#define MUTEX_UNLOCK(x) ReleaseMutex(&(x))
#define THREAD_ID GetCurrentThreadId()

#else /* !(defined(_WIN32) || defined(_WIN64)) */

#include <pthread.h>

#define _GNU_SOURCE
#include <unistd.h>

#define MUTEX_TYPE pthread_mutex_t
#define MUTEX_SETUP(x) pthread_mutex_init(&(x), NULL)
#define MUTEX_CLEANUP(x) pthread_mutex_destroy(&(x))
#define MUTEX_LOCK(x) pthread_mutex_lock(&(x))
#define MUTEX_UNLOCK(x) pthread_mutex_unlock(&(x))
#define THREAD_ID pthread_self()
#endif

#ifndef GO_OPENSSL_DEV
#define CRYPTO_LOCK 1
#endif

/* This array will store all of the mutexes available to OpenSSL. */
static MUTEX_TYPE *mutex_buf = NULL;

static void locking_function(int mode, int n, const char *file, int line)
{
	if (mode & CRYPTO_LOCK)
		MUTEX_LOCK(mutex_buf[n]);
	else
		MUTEX_UNLOCK(mutex_buf[n]);
}

static unsigned long id_function(void)
{
	return ((unsigned long)THREAD_ID);
}

int go_openssl_thread_setup(void)
{
	int i;

	mutex_buf = malloc(go_openssl_CRYPTO_num_locks() * sizeof(MUTEX_TYPE));
	if (!mutex_buf)
		return 0;
	for (i = 0; i < go_openssl_CRYPTO_num_locks(); i++)
		MUTEX_SETUP(mutex_buf[i]);
	go_openssl_CRYPTO_set_id_callback(id_function);
	go_openssl_CRYPTO_set_locking_callback(locking_function);
	return 1;
}
