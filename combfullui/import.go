package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"sync"
)

type import_item struct {
	length int
	prefix string
}

type import_stats struct {
	stacks_count  int
	tx_count      int
	merkle_count  int
	key_count     int
	decider_count int
}

func general_purpose_import(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	go func(done <-chan struct{}) {
		<-done
		r.Body.Close()
	}(r.Context().Done())

	var stats import_stats

	var bufr = bufio.NewReaderSize(r.Body, 2048)

	var err error

	var wg sync.WaitGroup

	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = bufr.ReadLine()
			buffer.Write(l)

			if !isPrefix {
				break
			}

			if err != nil {
				break
			}
		}

		if err == io.EOF {
			break
		} else if err == http.ErrBodyReadAfterClose {
			break
		}

		line := buffer.String()

		if len(line) < 3 {
			continue
		}

		if line[0] != '/' && line[0] != '\\' {
			continue
		}

		var import_data = []import_item{
			{156, wallet_header_net(WALLET_HEADER_STACK_DATA)},
			{1481, wallet_header_net(WALLET_HEADER_TX_RECV)},
			{1421, wallet_header_net(WALLET_HEADER_MERKLE_DATA)},
			{1357, wallet_header_net(WALLET_HEADER_WALLET_DATA)},
			{204, wallet_header_net(WALLET_HEADER_PURSE_DATA)},
		}

		for _, v := range import_data {

			if len(line) != v.length {
				continue
			}
			if line[0:len(v.prefix)] != v.prefix {
				continue
			}
			switch v.length {
			case 156:
				stats.stacks_count++
				wg.Add(1)
				go func() {
					stack_load_data_internal(DummyHttpWriter{}, line[12:])
					wg.Done()
				}()
			case 204:
				stats.decider_count++
				wg.Add(1)
				go func() {
					decider_load_data_internal(DummyHttpWriter{}, line[12:])
					wg.Done()
				}()
			case 1481:
				stats.tx_count++
				wg.Add(1)
				go func() {
					tx_receive_transaction_internal(DummyHttpWriter{}, line[9:])
					wg.Done()
				}()
			case 1421:
				stats.merkle_count++
				wg.Add(1)
				go func() {
					merkle_load_data_internal(DummyHttpWriter{}, line[13:])
					wg.Done()
				}()
			case 1357:
				stats.key_count++
				wg.Add(1)
				go func() {
					key_load_data_internal(DummyHttpWriter{}, line[13:])
					wg.Done()
				}()
			}
		}
	}

	wg.Wait()

	fmt.Fprintf(w, testnetColorBody()+`<a href="/">&larr; Back to home</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	fmt.Fprintln(w, stats)
}

func gui_import(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, testnetColorBody()+`
			<a href="/">&larr; Back to home</a><br />
			<form action="/import/general" method="post" enctype="multipart/form-data">
			    Select coin history file to import into wallet:
			    <input type="file" name="fileToUpload" id="fileToUpload">
			    <textarea style="display:none;" type="password" name="dataToUpload" id="dataToUpload"></textarea>
			    <input type="submit" value="Load Coin History" name="submit">
			</form>
			<script>

			function getReader() {
				var reader = new FileReader();
				  reader.onloadend = (function(file){
					document.getElementById('dataToUpload').innerHTML += this.result + "\r\n";
				  });
				  return reader;
			}

			document.addEventListener('paste', (event) => {

				var reader = getReader();

			      [...event.clipboardData.items].forEach((item, i) => {
			      // If dropped items aren't files, reject them
			      if (item.kind === 'file') {
				const file = item.getAsFile();
				document.getElementById('list').innerHTML += "<li>" + file.name + "</li>";
				reader.readAsText(file);
			      }
			    });
			})

			function dropHandler(ev) {

			  ev.preventDefault();

			 var reader = getReader();

			  if (ev.dataTransfer.items) {
			    // Use DataTransferItemList interface to access the file(s)
			    [...ev.dataTransfer.items].forEach((item, i) => {
			      // If dropped items aren't files, reject them
			      if (item.kind === 'file') {
				const file = item.getAsFile();
				document.getElementById('list').innerHTML += "<li>" + file.name + "</li>";
				reader.readAsText(file);
			      }
			    });
			  } else {
			    // Use DataTransfer interface to access the file(s)
			    [...ev.dataTransfer.files].forEach((file, i) => {
				document.getElementById('list').innerHTML += "<li>" + file.name + "</li>";
				reader.readAsText(file);
			    });
			  }
			}

			function dragOverHandler(ev) {
			  ev.preventDefault();
			}

			</script>
			<div
			  style="border: 5px solid blue; width: 90%; height: 11em;"
			  ondrop="dropHandler(event);"
			  ondragover="dragOverHandler(event);"><br /><br />
			  <p>Drag or paste one or more coin history files to this <i>drop zone</i>.</p>
			  <ul id="list"></ul>
			<br /><br /></div>


		</body></html>
	`)
}
