import debounce from './debounce.js';

document.addEventListener('DOMContentLoaded', () => {
  var canvas = document.querySelector("#fractal");
  if(!canvas) {
    return;
  }
  const dset = canvas.dataset;
  // var ctx = canvas.getContext('2d');
  Object.assign(dset, {
    centerx: -0.5,
    centery: 0,
    zoom: 200,
    resx: canvas.offsetWidth,
    resy: canvas.offsetHeight,
    offsetx: 0,
    offsety: 0,
    mousedown: false,
  });

  var draw = loadCanvasFractal(canvas, 1.5, 300);

  canvas.addEventListener('onresize', () => {
    Object.assign(dset, {
      resx: canvas.offsetWidth,
      resy: canvas.offsetHeight
    });
    draw();
  });
  var zoom = (amount) => {
    return (amt) => {
      dset.zoom = parseFloat(dset.zoom, 10) * (amount || amt);
      draw(true);
    };
  };
  document.querySelector("#zoom-in").onclick = zoom(1.25);
  document.querySelector("#zoom-out").onclick = zoom(1/1.25);

  const wheelFunc = (() => {
    var zoomfunc = zoom();
    var inc = 1.25;
    return ((e) => {
      e.preventDefault();
      zoomfunc(e.deltaY > 0 ? 1/inc : inc);
    });
  })();
  canvas.onmousewheel = wheelFunc;

  canvas.onmousedown = () => {
    if(dset.mousedown === "false") {
      dset.mousedown = true;
    }
  };
  canvas.onmouseup = () => {
    if(dset.mousedown === "true") {
      dset.mousedown = false;
      dset.centerx = parseFloat(dset.centerx, 10) -
        (dset.offsetx/dset.zoom + 1/(2*dset.zoom));
      dset.centery = parseFloat(dset.centery, 10) +
        (dset.offsety/dset.zoom + 1/(2*dset.zoom));
      dset.offsetx = 0;
      dset.offsety = 0;
      draw(true);
    }
  };
  canvas.onmousemove = (e) => {
    if(dset.mousedown === "true") {
      dset.offsetx = parseInt(dset.offsetx, 10) + e.movementX;
      dset.offsety = parseInt(dset.offsety, 10) + e.movementY;
      draw(false);
    }
  };
  draw(true);
});

function loadCanvasFractal(canvas, extra, delay) {
  var img;
  const ctx = canvas.getContext('2d');
  const debFetchFractal = debounce(async (params, putImage) => {
    const raw = await fetchFractal(params);
    img = new ImageData(params.resx, params.resy);
    for(let i = 0, j = 0; i < img.data.length; i += 4, j++) {
      img.data[i + 0] = raw[j] ? 0 : 255;
      img.data[i + 1] = raw[j] ? 0 : 255;
      img.data[i + 2] = raw[j] ? 0 : 255;
      img.data[i + 3] = 255;
    }
    putImage(img);
  }, delay);
  return async function(refetch) {
    const { centerx, centery, zoom, resx, resy, offsetx, offsety} = canvas.dataset;
    const [resxextra, resyextra] = [Math.ceil(resx*extra), Math.ceil(resy*extra)];
    const putImage = img => ctx.putImageData(img, offsetx-Math.ceil((resxextra-resx)/2), offsety-Math.ceil((resyextra-resy)/2));
    if (refetch || !img) {
      debFetchFractal({centerx, centery, zoom, resx: resxextra, resy: resyextra }, putImage);
    } else {
      putImage(img);
    }
  }
};

// params = { centerx, centery, zoom, resx, resy }
async function fetchFractal(params) {
  const urlParams = function() {
    const { centerx, centery, zoom, resx, resy } = params;
    return { centerx, centery, zoom, resx, resy };
  }();

  var url = new URL(window.location.href + 'fractal');
  Object.keys(params).forEach(key => url.searchParams.append(key, params[key]));

  var result = await fetch(url, { method:'GET', params});
  result = await result.json();
  return base64bin2bool(result.Image).slice(0, result.ResX*result.ResY);
}

function bin2bool(input) {
  let arr = [];
  for(let i = 0; i < 8; i++) {
    arr.push(!!(input&(1 << i)));
  }
  return arr;
};

function base64bin2bool(input, trueVal = true, falseVal = false) {
  var img = [];
  let encImg = atob(input);
  for (let i = 0; i < encImg.length;i++) {
    img.push(
      ...(bin2bool( encImg.charCodeAt(i)).reduce(
        (acc, item) => acc.concat(item ? trueVal : falseVal), []
      ))
    );
  }
  return img;
};