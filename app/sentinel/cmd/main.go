/*
|    Protect your secrets, protect your sensitive data.
:    Explore VMware Secrets Manager docs at https://vsecm.com/
</
<>/  keep your secrets... secret
>/
<>/' Copyright 2023-present VMware Secrets Manager contributors.
>/'  SPDX-License-Identifier: BSD-2-Clause
*/

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/akamensky/argparse"

	"github.com/vmware-tanzu/secrets-manager/app/sentinel/internal/safe"
	"github.com/vmware-tanzu/secrets-manager/core/crypto"
	entity "github.com/vmware-tanzu/secrets-manager/core/entity/data/v1"
)

func main() {
	parser := argparse.NewParser("safe", "Assigns secrets to workloads.")

	id, err := crypto.RandomString(8)
	if err != nil {
		id = "VSECMSENTINEL"
	}

	ctx, cancel := context.WithCancel(
		context.WithValue(context.Background(), "correlationId", &id),
	)

	defer cancel()

	list := parseList(parser)
	deleteSecret := parseDeleteSecret(parser)
	appendSecret := parseAppendSecret(parser)
	namespaces := parseNamespaces(parser)
	inputKeys := parseInputKeys(parser)
	backingStore := parseBackingStore(parser)
	workloadIds := parseWorkload(parser)
	secret := parseSecret(parser)
	template := parseTemplate(parser)
	format := parseFormat(parser)
	encrypt := parseEncrypt(parser)
	notBefore := parseNotBefore(parser)
	expires := parseExpires(parser)

	err = parser.Parse(os.Args)
	if err != nil {
		println(err.Error())
		println()
		printUsage(parser)
		return
	}

	if *list {
		if *encrypt {
			safe.Get(ctx, true)
			return
		}
		safe.Get(ctx, false)
		return
	}

	if *namespaces == nil || len(*namespaces) == 0 {
		*namespaces = []string{"default"}
	}

	if inputValidationFailure(workloadIds, encrypt, inputKeys, secret, deleteSecret) {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-c:
			println("Operation was cancelled.")
			cancel()
		}
	}()

	safe.Post(ctx, entity.SentinelCommand{
		WorkloadIds:        *workloadIds,
		Secret:             *secret,
		Namespaces:         *namespaces,
		BackingStore:       *backingStore,
		Template:           *template,
		Format:             *format,
		Encrypt:            *encrypt,
		DeleteSecret:       *deleteSecret,
		AppendSecret:       *appendSecret,
		SerializedRootKeys: *inputKeys,
		NotBefore:          *notBefore,
		Expires:            *expires,
	})
}
