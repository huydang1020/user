package db

import (
	"log"
	"testing"

	"github.com/huyshop/header/user"
	"github.com/huyshop/user/utils"
)

func Test_convertSlug(t *testing.T) {
	d := DB{}
	if err := d.ConnectDb("root:123456@tcp(localhost:3306)", "user"); err != nil {
		log.Println(err)
		return
	}

	listStore, err := d.ListStore(&user.StoreRequest{})
	if err != nil {
		log.Println(err)
		return
	}
	for _, store := range listStore {
		store.Slug = utils.ToSlug(store.Name)
		if err := d.UpdateStore(store, &user.Store{Id: store.Id}); err != nil {
			log.Println(err)
			return
		}
	}
	log.Println("len listStore", len(listStore))
}
