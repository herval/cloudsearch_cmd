package assets

import (
    "github.com/GeertJohan/go.rice"
)

func Static(name string) ([]byte, error) {
    page, err := rice.FindBox("../../static")
    if err != nil {
        return nil, err
    }

    return page.MustBytes("account_linked.html"), nil
}
