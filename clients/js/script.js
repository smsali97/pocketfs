var data = [{
    "id":1,"text":"Root node", "type": "folder","children": [
      {"id":2,"text":"Child node 1","type": "file"},
      {"id":3,"text":"Child node 2","type": "file"}
    ]
}]


$(function () {
    $('#jstree').jstree(
        {
            "types" : {
                "default" : {
                    "icon" : "glyphicon glyphicon-flash"
                },
                "file" : {
                  "icon" : "glyphicon glyphicon-file"
                },
                "folder" : {
                  "icon" : "glyphicon glyphicon-folder-open"
                }
            },
            'core' : {
                "themes" : { "stripes" : true,  },
                "multiple" : false, 
                'data' : data
            },
            'plugins' : ["types","wholerow"]
        }
    );
    $('#jstree').on("changed.jstree", function (e, data) {
      console.log(data.selected);
    });
    // 8 interact with the tree - either way is OK
    $('button').on('click', function () {
      $('#jstree').jstree('select_node', 'child_node_1');
    });
  });