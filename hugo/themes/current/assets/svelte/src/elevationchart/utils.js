function toRad(x) {
  return x * Math.PI / 180;
}

function haversineDistance(coords1, coords2) {
  var lon1 = coords1[0];
  var lat1 = coords1[1];

  var lon2 = coords2[0];
  var lat2 = coords2[1];

  var R = 6371000; // m

  var x1 = lat2 - lat1;
  var dLat = toRad(x1);
  var x2 = lon2 - lon1;
  var dLon = toRad(x2)
  var a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) *
    Math.sin(dLon / 2) * Math.sin(dLon / 2);
  var c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

  return R * c;
}

function geojsonFlatCoordinatesAndCoordTimes(feature) {
  if (feature.geometry.type == 'MultiLineString') {
    return [
      feature.geometry.coordinates.flat(1),
      feature.properties.coordTimes ? feature.properties.coordTimes.flat(1) : []
    ]
  } else if (feature.geometry.type == 'LineString') {
    return [
      feature.geometry.coordinates,
      feature.properties.coordTimes || [],
    ];
  }

  throw new Error(`unsupported geometry type (${feature.geometry.type})`);
}

function aggregateTransform(rawData) {
  const [ dataCoordinates, dataCoordTimes ] = geojsonFlatCoordinatesAndCoordTimes(rawData.features[0]);

  if (dataCoordTimes.length > 0 && dataCoordTimes.length != dataCoordinates.length) {
    throw new Error(`mismatch length between coordinates (${dataCoordinates.length}) and coordTimes (${dataCoordTimes.length})`);
  }

  let cumulativeDistance = 0;
  const dates = rawData.features[0].properties.coordTimes ? {
      maximum: dataCoordTimes[0],
      minimum: dataCoordTimes[0],
    } : undefined;
  const elevation = dataCoordinates[0].length > 2 ? {
      maximum: dataCoordinates[0][2],
      minimum: dataCoordinates[0][2],
      cumulativeGain: 0,
      cumulativeLoss: 0,
    } : undefined;
  const coordinates = {
    first: {
      latitude: dataCoordinates[0][1],
      longitude: dataCoordinates[0][0],
      elevation: dataCoordinates[0][2],
    },
    last: {
      latitude: dataCoordinates.slice(-1)[0][1],
      longitude: dataCoordinates.slice(-1)[0][0],
      elevation: dataCoordinates.slice(-1)[0][2],
    },
    latitude: {
      maximum: dataCoordinates[0][1],
      minimum: dataCoordinates[0][1],
    },
    longitude: {
      maximum: dataCoordinates[0][0],
      minimum: dataCoordinates[0][0],
    },
  }

  for (let datumIdx = 1; datumIdx < dataCoordinates.length; datumIdx++) {
    const datum = dataCoordinates[datumIdx];

    cumulativeDistance += haversineDistance(
      [ dataCoordinates[datumIdx-1][0], dataCoordinates[datumIdx-1][1] ],
      [ datum[0], datum[1] ],
    );

    if (datum[1] > coordinates.latitude.maximum) {
      coordinates.latitude.maximum = datum[1];
    }

    if (datum[1] < coordinates.latitude.minimum) {
      coordinates.latitude.minimum = datum[1];
    }

    if (datum[0] > coordinates.longitude.maximum) {
      coordinates.longitude.maximum = datum[0];
    }

    if (datum[0] < coordinates.longitude.minimum) {
      coordinates.longitude.minimum = datum[0];
    }
    
    if (dates) {
      const datumCoordTime = dataCoordTimes[datumIdx];

      if (datumCoordTime > dates.maximum) {
        dates.maximum = datumCoordTime;
      }

      if (datumCoordTime < dates.minimum) {
        dates.minimum = datumCoordTime;
      }
    }

    if (elevation) {
      if (datum[2] > elevation.maximum) {
        elevation.maximum = datum[2];
      }

      if (datum[2] < elevation.minimum) {
        elevation.minimum = datum[2];
      }

      const elevationDiff = datum[2] - dataCoordinates[datumIdx-1][2];

      if (elevationDiff < 0) {
        elevation.cumulativeLoss -= elevationDiff;
      } else {
        elevation.cumulativeGain += elevationDiff;
      }
    }
  }

  return {
    cumulativeDistance,
    dates,
    elevation,
    coordinates,
  };
}

export { aggregateTransform };

function chartTransform(rawData, options) {
  const reducePoints = options.reducePoints;
  const styleColorDefault = options.styleColorDefault || '#fafafa'; // text-stone-50
  const styleColorModerate = options.styleColorDefault || '#d6d3d1'; // text-stone-300
  const styleColorDifficult = options.styleColorDefault || '#a8a29e'; // text-stone-400
  const styleColorStrenuous = options.styleColorDefault || '#57534e'; // text-stone-600

  //

  const [ dataCoordinates, dataCoordTimes ] = geojsonFlatCoordinatesAndCoordTimes(rawData.features[0]);

  const data = dataCoordinates.map((v, vIdx) => ({
    date: dataCoordTimes[vIdx],
    distance: 0,
    lnglat: [v[1], v[0]],
    elevation: v[2],
    styleColor: styleColorDefault,
  }));

  let distance = 0;

  for (let datumIdx = 1; datumIdx < data.length; datumIdx++) {
    const diffdist = haversineDistance(data[datumIdx-1].lnglat, data[datumIdx].lnglat);

    distance += diffdist;
    data[datumIdx].distance = distance;
  }

  // reduce

  let reducedData = data;

  if (reducePoints && reducePoints < (data.length * 1.1)) {
    const reducedStepDistance = distance / reducePoints;

    reducedData = [data[0]];

    let reducedDistanceThreshold = data[0].distance + reducedStepDistance;
    let reducedDatum = {
      ...data[1],
      reducedPoints: 1,
    };

    for (const datum of data.slice(2, -1)) {        
      if (datum.distance >= reducedDistanceThreshold) {
        reducedData.push(reducedDatum);

        reducedDatum = {
          ...datum,
          reducedPoints: 1,
        };
        reducedDistanceThreshold = datum.distance + reducedStepDistance;

        continue;
      }

      reducedDatum.date = datum.date;
      reducedDatum.distance = datum.distance;
      reducedDatum.lnglat = datum.lnglat;
      reducedDatum.elevation = datum.elevation;
      reducedDatum.reducedPoints++;
    }

    reducedData.push(reducedDatum, data.slice(-1)[0]);
  }

  // styleColor

  for (let datumIdx = 0; datumIdx < reducedData.length; datumIdx++) {
    const fromDatum = reducedData[Math.max(0, datumIdx - 2)];
    const thruDatum = reducedData[Math.min(reducedData.length - 1, datumIdx + 1)];

    if (!fromDatum || !thruDatum) {
      continue;
    }

    const averageRate = (thruDatum.elevation - fromDatum.elevation) / (thruDatum.distance - fromDatum.distance);

    if (averageRate > 0.16) {
      reducedData[datumIdx].styleColor = styleColorStrenuous;
    } else if (averageRate > 0.12) {
      reducedData[datumIdx].styleColor = styleColorDifficult;
    } else if (averageRate > 0.06) {
      reducedData[datumIdx].styleColor = styleColorModerate;
    }

    reducedData[datumIdx].elevationRateAverage = averageRate;
  }

  return reducedData;
}

export { chartTransform };
