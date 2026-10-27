// Package storage guarda e lê os arquivos no Cloudflare R2, pela API S3 (minio-go).
//
// O arquivo nunca é carregado inteiro na memória: o envio recebe um io.Reader com tamanho
// conhecido (sem tamanho, o minio-go segura partes de 16 MB na RAM), e a leitura devolve um
// objeto que busca do R2 aos pedaços, conforme a resposta é escrita.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrNotFound é a chave que não existe no bucket.
var ErrNotFound = errors.New("storage: not found")

// Settings são as credenciais do R2 (lib/infra/config.Storage).
type Settings struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// Object é um arquivo lido do R2, pronto para o http.ServeContent: lê, pula para qualquer
// ponto (pedido de faixa de um PDF ou vídeo) e fecha.
type Object struct {
	Content      io.ReadSeekCloser
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
}

// Bucket é um bucket do R2.
type Bucket struct {
	client *minio.Client
	name   string
}

// Open conecta ao bucket. Não faz chamada de rede: a primeira acontece no primeiro arquivo.
func Open(settings Settings) (*Bucket, error) {
	endpoint, err := url.Parse(settings.Endpoint)
	if err != nil || endpoint.Host == "" {
		return nil, fmt.Errorf("storage: R2_ENDPOINT must be a URL like https://<account>.r2.cloudflarestorage.com, got %q", settings.Endpoint)
	}

	client, err := minio.New(endpoint.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.AccessKeyID, settings.SecretAccessKey, ""),
		Secure: endpoint.Scheme != "http",
		// O R2 não tem regiões: "auto" é o que ele aceita na assinatura.
		Region: "auto",
	})
	if err != nil {
		return nil, fmt.Errorf("storage: open client: %w", err)
	}

	return &Bucket{client: client, name: settings.Bucket}, nil
}

// Put grava content (com size bytes) em key.
func (bucket *Bucket) Put(ctx context.Context, key string, content io.Reader, size int64, contentType string) error {
	_, err := bucket.client.PutObject(ctx, bucket.name, key, content, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("storage: put %s: %w", key, err)
	}

	return nil
}

// Get abre key para leitura. Quem chama fecha o Content.
func (bucket *Bucket) Get(ctx context.Context, key string) (Object, error) {
	object, err := bucket.client.GetObject(ctx, bucket.name, key, minio.GetObjectOptions{})
	if err != nil {
		return Object{}, notFoundOr(err)
	}

	// O Stat é o que faz o pedido ao R2: é aqui que "não existe" aparece.
	info, err := object.Stat()
	if err != nil {
		_ = object.Close()
		return Object{}, notFoundOr(err)
	}

	return Object{
		Content:      object,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
	}, nil
}

// DeleteFolder apaga tudo o que começa com prefix — as versões de um artefato apagado.
func (bucket *Bucket) DeleteFolder(ctx context.Context, prefix string) error {
	objects := bucket.client.ListObjects(ctx, bucket.name, minio.ListObjectsOptions{Prefix: prefix, Recursive: true})

	for failure := range bucket.client.RemoveObjects(ctx, bucket.name, objects, minio.RemoveObjectsOptions{}) {
		if failure.Err != nil {
			return fmt.Errorf("storage: delete %s: %w", failure.ObjectName, failure.Err)
		}
	}

	return nil
}

func notFoundOr(err error) error {
	if minio.ToErrorResponse(err).Code == minio.NoSuchKey {
		return ErrNotFound
	}

	return err
}
