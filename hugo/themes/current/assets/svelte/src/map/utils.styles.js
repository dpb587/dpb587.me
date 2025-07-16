import { Icon, Stroke, Style } from 'ol/style';
import { drawRoundRect } from "./utils.styles.rounded";

const defaultFeatureStyles = {
  Polygon: new Style({
    stroke: new Stroke({
      lineDash: [6, 8],
      color: 'rgba(12,74,110	,0.5)', // text-sky-900
      width: 4,
    }),
  }),
  LineString: new Style({
    stroke: new Stroke({
      // lineDash: [6, 8],
      color: 'rgba(12,74,110	,0.9)', // text-sky-900
      width: 4,
    }),
  }),
  MultiLineString: new Style({
    stroke: new Stroke({
      // lineDash: [6, 8],
      color: 'rgba(12,74,110	,0.9)', // text-sky-900
      width: 4,
    }),
  }),
};

export { defaultFeatureStyles };

const customFeatureCache = {} // TODO leaky, racy, prop-collision

const getFeatureStyle = function (featureStyles, feature) {
  let s = featureStyles[feature.getGeometry().getType()]

  const fi = feature.get('image')
  if (fi) {
    if (customFeatureCache[fi]) {
      return customFeatureCache[fi]
    }

    s = s.clone()

    const img = new Image()
    // img.crossOrigin = 'anonymous';
    img.onload = function () {
      const canvas = document.createElement('canvas');
      canvas.width = 90;
      canvas.height = 90;

      let ctx = canvas.getContext('2d');

      ctx.strokeStyle = "rgba(255, 255, 255, 1)";
      ctx.lineWidth = 6
      ctx.shadowColor = 'rgba(100, 100, 100, 0.6)';
      ctx.shadowBlur = 5;
      ctx.shadowOffsetX = 1;
      ctx.shadowOffsetY = 1;
      drawRoundRect(ctx, 6, 6, 70, 70, 6, true, true)

      ctx.shadowColor = null
      ctx.shadowBlur = 0
      ctx.shadowOffsetX = 0
      ctx.shadowOffsetY = 0

      ctx.drawImage(img, 0, 0, img.width, img.height, 6, 6, 70, 70);

      ctx.strokeStyle = "rgba(255, 255, 255, 1)";
      ctx.lineWidth = 4
      drawRoundRect(ctx, 6, 6, 70, 70, 6, false, true);

      s.setImage(new Icon({
        anchor: [ 0.48, 0.48 ],
        anchorXUnits: 'fraction',
        anchorYUnits: 'fraction',
        img: canvas,
        imgSize: [canvas.width, canvas.height],
        scale: 0.5
      }))

      feature.setStyle(s)

      customFeatureCache[fi] = s
    }
    
    img.src = fi;

    s.setImage(null)

    // s.setImage(new Icon({
    //   anchor: [0.25, 0.25],
    //   anchorXUnits: 'fraction',
    //   anchorYUnits: 'fraction',
    //   size: [ 96, 96],
    //   src: 'https://img.icons8.com/plasticine/100/000000/place-marker.png',
    //   scale: 0.5,
    //   // imgSize: 24,
    // }))
  // } else {
  //   const st = s.getText()
  //   if (st) {
  //     st.setText(feature.get('name'))
  //   }
  }

  return s
}

export { getFeatureStyle }
