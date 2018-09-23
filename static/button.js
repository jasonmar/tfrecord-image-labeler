// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

var GoogleGreen = '#3cba54';
var GoogleYellow = '#f4c20d';
var GoogleRed = '#db3236';
var GoogleBlue = '#4885ed';

// front left bottom (light, mid, dark)
var blue = ['#3e80f6', '#2874e2', '#2068d1'];
var grey = ['#cfd1d2', '#babcbe', '#a5a7aa'];
var neutral = ['#a4a6a9', '#909295', '#808082'];
var green = ['#2baa5e', '#249b54', '#1a8741'];
var purple = ['#c95ac5', '#b74eb7', '#a53fa5'];
var orange = ['#dd4d31', '#cc402e', '#bf3025'];
var fb = ['#0066a8', '#0080c4', '#004a75'];
var tw = ['#00abe0', '#89d9f7', '#0080ab'];

function mkLabelCanvas(canvasId) {
  var x = {
    canvas: new fabric.Canvas(canvasId),
    bbox: BBox(),
    pt1: LabelCanvas.Circle(0, 0),
    pt2: LabelCanvas.Circle(0, 0),
    image: null,
    img: null
  }
  LabelCanvas.reset(x);
  return x;
}

function BBox() {
  var x = {
    top: LabelCanvas.Line(0,0,0,0),
    left: LabelCanvas.Line(0,0,0,0),
    right: LabelCanvas.Line(0,0,0,0),
    bottom: LabelCanvas.Line(0,0,0,0)
  }
  return x;
}

function mkImage(filename, labelId, labelDesc, format) {
  return {
    filename: filename,
    labelId: labelId,
    labelDesc: labelDesc,
    format: format
  }
}

class LabelCanvas {
  constructor() {
  }

  static setHandlers(x){
    x.canvas.on({
      'mouse:up': function (event) {
        var ptr = x.canvas.getPointer(event.e);
        if (!x.pt1.visible) {
          x.pt1.top = ptr.y;
          x.pt1.left = ptr.x;
          x.pt1.visible = true;
          x.pt1.bringToFront();
          x.canvas.renderAll();
        } else if (!x.pt2.visible) {
          if (ptr.x != x.pt1.left && ptr.y != x.pt1.top) {
            x.pt2.top = ptr.y;
            x.pt2.left = ptr.x;
            x.pt2.visible = true;
            x.pt2.bringToFront();
            LabelCanvas.update(x);
          }
        } else if (x.pt1.visible && x.pt2.visible) {
          LabelCanvas.reset(x);
        }
      },
      'mouse:down': function (event) {
        var ptr = x.canvas.getPointer(event.e);
        if (!x.pt1.visible) {
          x.pt1.top = ptr.y;
          x.pt1.left = ptr.x;
          x.pt1.visible = true;
          x.pt1.bringToFront();
          x.canvas.renderAll();
        }
      },
      'object:moving': function(e) { LabelCanvas.update(x); }
    });
  }

  static setImg(x, filename, labelId, labelDesc, w, h, format) {
    // label data
    x.image = mkImage(filename, labelId, labelDesc, format);

    // canvas object
    var onLoad = function(img) {
      x.img = img;
      var oDim = img.getOriginalSize();
      x.image.width = oDim.width;
      x.image.height = oDim.height;
      // fit image to canvas
      x.img.scaleX = (w/oDim.width);
      x.img.scaleY = (h/oDim.height);
      LabelCanvas.reset(x);
    }
    var imgOpts = {
      hasControls: false,
      hasBorders: false,
      selectable: false,
      hoverCursor: 'default'
    }
    fabric.Image.fromURL(filename, onLoad, imgOpts);
  }

  static reset(x) {
    x.canvas.clear();
    
    x.canvas.add(x.bbox.top);
    x.canvas.add(x.bbox.left);
    x.canvas.add(x.bbox.right);
    x.canvas.add(x.bbox.bottom);

    x.canvas.add(x.pt1);
    x.canvas.add(x.pt2);

    if (x.img != null) {
      x.canvas.add(x.img);
      x.img.sendToBack();
    }

    x.bbox.top.visible = false;
    x.bbox.left.visible = false;
    x.bbox.right.visible = false;
    x.bbox.bottom.visible = false;

    x.pt1.visible = false;
    x.pt2.visible = false;

    x.pt1.bringToFront();
    x.pt2.bringToFront();

    x.canvas.renderAll();
  }

  static update(x) {
    x.bbox.top.visible = true;
    x.bbox.left.visible = true;
    x.bbox.right.visible = true;
    x.bbox.bottom.visible = true;

    x.bbox.top.set({
      x1: x.pt1.left,
      y1: x.pt1.top,
      x2: x.pt2.left,
      y2: x.pt1.top
    })

    x.bbox.left.set({
      x1: x.pt1.left,
      y1: x.pt1.top,
      x2: x.pt1.left,
      y2: x.pt2.top
    })

    x.bbox.right.set({
      x1: x.pt2.left,
      y1: x.pt1.top,
      x2: x.pt2.left,
      y2: x.pt2.top
    })

    x.bbox.bottom.set({
      x1: x.pt1.left,
      y1: x.pt2.top,
      x2: x.pt2.left,
      y2: x.pt2.top
    })

    x.canvas.renderAll();
  }

  static getBB(x) {
    if (x.bbox.top.visible && x.image != null) {
      return {
        'image/height': x.image.height,
        'image/width': x.image.width,
        'image/filename': x.image.filename,
        'image/source_id': x.image.filename,
        'image/format': x.image.format,
        'image/object/bbox/xmin': Math.min(x.pt1.left, x.pt2.left) / x.image.width,
        'image/object/bbox/xmax': Math.max(x.pt1.left, x.pt2.left) / x.image.width,
        'image/object/bbox/ymin': Math.min(x.pt1.top, x.pt2.top) / x.image.height,
        'image/object/bbox/ymax': Math.max(x.pt1.top, x.pt2.top) / x.image.height,
        'image/object/class/text': x.image.labelDesc,
        'image/object/class/label': x.image.labelId
      }
    } else {
      return null;
    }
  }

  static Line(x1, y1, x2, y2) {
    var l = new fabric.Line([x1, y1, x2, y2], {
      stroke: '#f4c20d',
      strokeWidth: 2,
      hasControls: false,
      hasBorders: false,
      selectable: false,
      visible: false,
      hoverCursor: 'default'
    });
    return l;
  }

  static Circle(x, y) {
    var c = new fabric.Circle({
      left: x,
      top: y,
      strokeWidth: 2,
      radius: 3,
      stroke: '#3cba54',
      fill: '#db3236',
      originX: 'center',
      originY: 'center',
      hasControls: false,
      hasBorders: false,
      selectable: false,
      visible: false,
      hoverCursor: 'default'
    });
    return c;
  }

  static Rect(x,y,w,h) {
    var rect = new fabric.Rect({
      left: x,
      top: y,
      fill: 'rgba(0,0,0,0)',
      originX: 'top',
      originY: 'left',
      width: w,
      height: h,
      stroke: '#f4c20d',
      strokeWidth: 2,
      hasControls: false,
      hasBorders: false,
      selectable: false,
      visible: false,
      hoverCursor: 'default'
    });
    return rect;
  }
}


function mkbutton(canvas, left, top, width, height, colors, depth, depth1, txt) {

  var frStartTop = top + height;
  
  var lr = new fabric.Rect({
    originX: 'left',
    originY: 'bottom',
    left: left,
    top: top + height + depth,
    width: depth,
    height: height,
    fill: colors[1],
    stroke: colors[1],
    strokeWidth: 0,
    skewY: -45
  });
  
  var br = new fabric.Rect({
    originX: 'left',
    originY: 'bottom',
    left: lr.left,
    top: lr.top,
    width: width,
    height: depth,
    fill: colors[2],
    stroke: colors[2],
    strokeWidth: 0,
    skewX: -45
  });
  
  var fr = new fabric.Rect({
    originX: 'left',
    originY: 'bottom',
    left: lr.left + depth,
    top: br.top - depth,
    width: width,
    height: height,
    fill: colors[0],
    stroke: colors[0],
    strokeWidth: 0
  });

  var text = new fabric.Text(txt, {
    originX: 'center',
    originY: 'center',
    left: fr.left + (fr.width / 2),
    top: fr.top - (fr.height / 2),
    fill: '#ffffff',
    fontSize: 28,
    fontFamily: 'Poppins',
    evented: false
  });

  setDefaults(text);
  setDefaults(lr);
  setDefaults(br);
  setDefaults(fr);

  var group = new fabric.Group([lr, br, fr, text], {
    originX: 'left',
    originY: 'bottom'
  });
  setDefaults(group);

  var d0 = 280;
  var sync = function() {
    fr.top = br.top - br.height;
    fr.left = lr.left + lr.width;
    text.left = fr.left + (fr.width / 2);
    text.top = fr.top - (fr.height / 2);
    canvas.renderAll.bind(canvas)();
  }
  var animateFn = function() {
    lr.animate('width', depth1, {
      onChange: sync,
      duration: d0,
      easing: fabric.util.ease.easeOutBounce
    });
    br.animate('height', depth1, {
      duration: d0,
      easing: fabric.util.ease.easeOutBounce
    });
  }

  var d1 = 600;
  var animateUpFn = function() {
    lr.animate('width', depth, {
      duration: d1,
      easing: fabric.util.ease.easeOutBounce
    });
    br.animate('height', depth, {
      onChange: sync,
      duration: d1,
      easing: fabric.util.ease.easeOutBounce
    });
  }

  return {
    group: group,
    animateFn: animateFn,
    animateUpFn: animateUpFn
  }
}

function setDefaults(o) {
  o.hasControls = false;
  o.hasBorders = false;
  o.selectable = false;
  o.hoverCursor = 'default';
}

function createButton(canvasId, colors, w, h, d, d1, onPush) {
  var canvas = new fabric.Canvas(canvasId);
  var btn = mkbutton(canvas, 0, 0, w - d, h - d, colors, d, d1, 'PUSH');
  var onMouseDown = function(e) {
    btn.animateFn();
    onPush();
  }
  var onMouseUp = function(e) { btn.animateUpFn(); }

  canvas.on({
    'mouse:up': onMouseUp,
    'mouse:down': onMouseDown
  });
  canvas.add(btn.group);
  canvas.renderAll();
}

function postBBox(bbox) {
  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/label', true);
  xhr.setRequestHeader('Content-Type', 'application/json; charset=UTF-8');
  xhr.send(JSON.stringify(bbox));
}

function getImg(labeler) {
  var xhr = new XMLHttpRequest();
  var onLoadEnd = function(e){
    var img = xhr.response;
    var uri = img['uri'];
    LabelCanvas.setImg(labeler, uri, 1, 'class1', 320, 320, 'jpg');
  }
  xhr.open('GET', '/image', true);
  xhr.addEventListener("loadend", onLoadEnd);
  xhr.responseType = "json";
  xhr.send();
}

window.onload = function(e){
  var labeler = mkLabelCanvas('label');
  getImg(labeler);
  LabelCanvas.setHandlers(labeler);
  var onPush = function() {
    var bbox = LabelCanvas.getBB(labeler);
    if (bbox != null) {
      postBBox(bbox);
    }
    getImg(labeler);
  }
  document.body.onkeyup = function(e) { if (e.keyCode == 32) { onpush(); } }
  createButton('button', green, 320, 80, 18, 6, onPush);
}
