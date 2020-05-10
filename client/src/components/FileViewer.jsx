import React from 'react'
import Moment from 'moment'
import FileBrowser, {Icons, FileRenderers, FolderRenderers, Groupers}  from 'react-keyed-file-browser'
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import Snackbar from '@material-ui/core/Snackbar';
import { saveAs } from 'file-saver';

const useStyles = makeStyles(theme => ({
  root: {
    '& > *': {
      margin: theme.spacing(1),
    },
  },
  input: {
    display: 'none',
  },
  blue: {
    color: 'lightblue',
  }
}));



class FileViewer extends React.Component {
  ws = new WebSocket('ws://localhost:8000/ws')
  
  componentDidMount() {
    this.timer = setInterval(() => this.fetchFiles(), 5000);  
  }

  componentWillMount(){
    this.ws.onopen = () => {
      // on connecting, do nothing but log it to the console
      console.log('connected to proxy :)')
      }

      this.ws.onmessage = evt => {
      // listen to data sent from the websocket server
      const message = evt.data
          if (message.includes("PING")) {
              var serverIp = JSON.parse(message.replace('PING ',''))

              let newServers = this.state.servers.filter( e => 
                e['ip'] !== serverIp["ip"] 
              )
              newServers.push(serverIp)
              newServers.sort((a, b) => {
                return a.latency - b.latency;
              });
              this.setState({...this.state, servers: newServers})
              console.log(this.state.servers)
            }

      }

      this.ws.onclose = () => {
        console.log('disconnected from proxy :(')
        this.ws = new WebSocket('ws://localhost:8000/ws')
      }
  }

  componentWillUnmount() {
    this.timer = null;
  }  

  
  state = { 
    snackBarError: false,
    snackBarMessage: "",
    files: [],
    servers: []
  }

  fetchFiles = () => {
    this.setState({...this.state, isFetching: true})
    if (this.state.servers.length === 0) {
        this.setState({...this.state, snackBarError: true, snackBarMessage: 'Unable to get directory.. no servers online :('})
        return
    } else {
      this.setState({...this.state, snackBarError: false, snackBarMessage: ''})
    }
    
    fetch('http://' + this.state.servers[0].ip + ":8080/directory")
      .then(response => response.json())
      .then(result => {
        result.forEach( r => {
            r["modified"] = Moment(r["modified"])
        })
        this.setState({...this.state, files: result, isFetching: false})})
      .catch(e => {
        this.state.servers.shift()
        console.log(e)
      });
  }

  handleDownloadFile = (files) => {

    files.forEach( file => {
    saveAs('http://' + this.state.servers[0].ip  + ':8080/download?path=' + file,file)
    return
    }
    );
  }

  handleClose = (event, reason) => {
    if (reason === 'clickaway') {
      return;
    }
      this.setState( state => {
        state.snackBarError = false;
        state.snackBarMessage = "";
      } )
  };

  handleCreateFolder = (key) => {
      const requestOptions = {
        method: 'POST',
    };


    fetch('http://' + this.state.servers[0].ip  + ':8080/directory?path=' + key.slice(0, -1), requestOptions)
        .then(response => {
            const status = response.headers.get("status");
          console.log(response)
            if (status !== 200) {
              
              response.text().then(r => {
                    this.setState( s => {
                          s.snackBarError = true;
                          s.snackBarMessage = r;
                    })
              })
            }
        })
  }
  handleCreateFiles = (files, prefix) => {
    console.log(files)
    console.log(prefix)

    files.forEach( file => {



    const formData = new FormData();
    formData.append('myFile', file)
    if (prefix !== "") prefix = prefix.slice(0,-1)

    const requestOptions = {
      method: 'POST',
      body: formData 
    };

    fetch('http://' + this.state.servers[0].ip  + ':8080/upload?path=' + prefix, requestOptions)
    .then(response => {
        const status = response.headers.get("status");
      console.log(response)
        if (status !== 200) {
          response.text().then(r => {
                this.setState( s => {
                      s.snackBarError = true;
                      s.snackBarMessage = r;
                })
          })
        }
    })

  })
  }
  handleRenameFolder = (oldKey, newKey) => {
    this.setState(state => {
      const newFiles = []
      state.files.map((file) => {
        if (file.key.substr(0, oldKey.length) === oldKey) {
          newFiles.push({
            ...file,
            key: file.key.replace(oldKey, newKey),
            modified: +Moment(),
          })
        } else {
          newFiles.push(file)
        }
      })
      state.files = newFiles
      return state
    })
  }
  handleRenameFile = (oldKey, newKey) => {
    this.setState(state => {
      const newFiles = []
      state.files.map((file) => {
        if (file.key === oldKey) {
          newFiles.push({
            ...file,
            key: newKey,
            modified: +Moment(),
          })
        } else {
          newFiles.push(file)
        }
      })
      state.files = newFiles
      return state
    })
  }
  handleDeleteFolder = (key) => {
    const requestOptions = {
      method: 'DELETE',
  };
  key = key[0].slice(0, -1)
  console.log(key)

  fetch('http://' + this.state.servers[0].ip  + ':8080/directory?path=' + key, requestOptions)
      .then(response => {
          const status = response.headers.get("status");
        console.log(response)
          if (status !== 200) {
            response.text().then(r => {
                  this.setState( s => {
                        s.snackBarError = true;
                        s.snackBarMessage = r;
                  })
            })
          }
      })  }
  handleDeleteFile = (fileKey) => {
    const requestOptions = {
      method: 'POST',
    };
    // TODO: Make this for multiple files
    var prefix = fileKey[0]
    fetch('http://' + this.state.servers[0].ip  + ':8080/remove?path=' + prefix, requestOptions)
    .then(response => {
        const status = response.headers.get("status");
      console.log(response)
        if (status !== 200) {
          response.text().then(r => {
                this.setState( s => {
                      s.snackBarError = true;
                      s.snackBarMessage = r;
                })
          })
        }
    })
  }

  
  render() {

    // const classes = useStyles();

    return (

      <Grid container spacing={3}>
        <Grid item xs={12}    container spacing={0}
  direction="column"
  alignItems="center"
  justify="center">
      <img src="lifting-1TB.png"></img>
      <Typography variant="h7" component="h7" gutterBottom>
       <em>{this.state.servers.length} Pocket Gopher{this.state.servers.length !== 1 ? 's are' : ' is'} available to serve you files!</em>
      </Typography>
      <Typography variant="h3" component="h2" gutterBottom>
        <em>PocketFS -</em> Distributed File System
      </Typography>
      <Snackbar open={this.state.snackBarError} autoHideDuration={6000} onClose={this.handleClose}>
        <Typography variant="h5" component="h5" gutterBottom>  
        {this.state.snackBarMessage}
        </Typography>
      </Snackbar>
          <Paper className>
      <FileBrowser
        files={this.state.files}
        icons={Icons.FontAwesome(4)}
        onDownloadFile={this.handleDownloadFile}
        onCreateFolder={this.handleCreateFolder}
        onCreateFiles={this.handleCreateFiles}
        onMoveFolder={this.handleRenameFolder}
        onMoveFile={this.handleRenameFile}
        onRenameFolder={this.handleRenameFolder}
        // onRenameFile={this.handleRenameFile}
        onDeleteFolder={this.handleDeleteFolder}
        onDeleteFile={this.handleDeleteFile}
      />
      </Paper>
        </Grid>
        {/* <Grid item xs={6}>
          <Paper className></Paper>
       </Grid> */}
       </Grid>
    )
  }
}

export default FileViewer;