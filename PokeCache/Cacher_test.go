package pokecache

import (
	"testing"
    "time"
)

 func TestAddGet(t *testing.T){
     c := NewCache(5)
     c.Add("siddharth", []byte("9121336622"))
     val,err := c.Get("siddharth")
     if err != nil{
         t.Errorf("%v\n",err)
     }
     t.Logf("The value for %s : %s \n","siddharth",string(val))
     val,err = c.Get("Nivruth")
     if err == nil{
         t.Errorf("Nivruth found in cache with %v",val)
     }
     t.Logf("%v\n",err)
 }

 func TestCacheEntryDeleter(t *testing.T){
     c:= NewCache(5)
     c.Add("siddharth", []byte("9121336622"))
     time.Sleep(2 * time.Second)
     val,err := c.Get("siddharth")
     if err != nil{
         t.Errorf("%v\n",err)
     }
     t.Logf("The value for %s : %s \n","siddharth",string(val))
     time.Sleep(c.cacheEntryDeletionTime - (2*time.Second))
     val,err = c.Get("siddharth")
     if err == nil{
        t.Fatalf("Found entry for siddharth, with val %v when it should've been deleted",string(val))
     }
     t.Log("Couldn't find entry for Siddhart. Entry successfully deleted")
 }
