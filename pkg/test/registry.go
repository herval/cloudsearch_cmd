package test

import "github.com/herval/cloudsearch/pkg"

func DefaultRegistry() *cloudsearch.Registry {
    c := cloudsearch.NewRegistry()

    c.RegisterContentTypes(
        cloudsearch.Document,
        cloudsearch.Email,
        cloudsearch.File,
        cloudsearch.Folder,
        cloudsearch.Image,
        cloudsearch.Video,
    )

    return c
}
