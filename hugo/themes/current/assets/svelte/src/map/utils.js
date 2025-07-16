import { getFeatureStyle } from "./utils.styles";
import { Fill, Stroke, Style, Text } from "ol/style";
import CircleStyle from "ol/style/Circle";
import { LineString, Point, Polygon } from "ol/geom";

const circleDistanceMultiplier = 1;
const circleFootSeparation = 28;
const circleStartAngle = Math.PI / 2;

const convexHullFill = new Fill({
  color: 'rgba(255, 153, 0, 0.4)',
});
const convexHullStroke = new Stroke({
  color: 'rgba(204, 85, 0, 1)',
  width: 1.5,
});
const outerCircleFill = new Fill({
  color: 'rgba(255, 153, 102, 0.3)',
});
const innerCircleFill = new Fill({
  color: 'rgba(255, 165, 0, 0.7)',
});
const textFill = new Fill({
  color: '#fff',
});
const textStroke = new Stroke({
  color: 'rgba(0, 0, 0, 0.6)',
  width: 3,
});
const innerCircle = new CircleStyle({
  radius: 14,
  fill: innerCircleFill,
});
const outerCircle = new CircleStyle({
  radius: 20,
  fill: outerCircleFill,
});

export {
  circleDistanceMultiplier,
  circleFootSeparation,
  circleStartAngle,
  convexHullFill,
  convexHullStroke,
  outerCircleFill,
  innerCircleFill,
  textFill,
  textStroke,
  innerCircle,
  outerCircle,
}


/**
 * Single feature style, users for clusters with 1 feature and cluster circles.
 * @param {Feature} clusterMember A feature from a cluster.
 * @return {Style} An icon style for the cluster member's location.
 */
 function clusterMemberStyle(featureStyles, clusterMember) {
  return getFeatureStyle(featureStyles, clusterMember)
  // return new Style({
  //   geometry: clusterMember.getGeometry(),
  //   image: clusterMember.get('LEISTUNG') > 5 ? darkIcon : lightIcon,
  // });
}

let clickFeature, clickResolution;
/**
 * Style for clusters with features that are too close to each other, activated on click.
 * @param {Feature} cluster A cluster with overlapping members.
 * @param {number} resolution The current view resolution.
 * @return {Style} A style to render an expanded view of the cluster members.
 */
function clusterCircleStyle(featureStyles, cluster, resolution) {
  if (cluster !== clickFeature || resolution !== clickResolution) {
    return;
  }
  const clusterMembers = cluster.get('features');
  const centerCoordinates = cluster.getGeometry().getCoordinates();
  return generatePointsCircle(
    clusterMembers.length,
    cluster.getGeometry().getCoordinates(),
    resolution
  ).reduce((styles, coordinates, i) => {
    const point = new Point(coordinates);
    const line = new LineString([centerCoordinates, coordinates]);
    styles.unshift(
      new Style({
        geometry: line,
        stroke: convexHullStroke,
      })
    );
    styles.push(
      clusterMemberStyle(
        featureStyles,
        new Feature({
          ...clusterMembers[i].getProperties(),
          geometry: point,
        })
      )
    );
    return styles;
  }, []);
}

/**
 * From
 * https://github.com/Leaflet/Leaflet.markercluster/blob/31360f2/src/MarkerCluster.Spiderfier.js#L55-L72
 * Arranges points in a circle around the cluster center, with a line pointing from the center to
 * each point.
 * @param {number} count Number of cluster members.
 * @param {Array<number>} clusterCenter Center coordinate of the cluster.
 * @param {number} resolution Current view resolution.
 * @return {Array<Array<number>>} An array of coordinates representing the cluster members.
 */
function generatePointsCircle(count, clusterCenter, resolution) {
  const circumference =
    circleDistanceMultiplier * circleFootSeparation * (2 + count);
  let legLength = circumference / (Math.PI * 2); //radius from circumference
  const angleStep = (Math.PI * 2) / count;
  const res = [];
  let angle;

  legLength = Math.max(legLength, 35) * resolution; // Minimum distance to get outside the cluster icon.

  for (let i = 0; i < count; ++i) {
    // Clockwise, like spiral.
    angle = circleStartAngle + i * angleStep;
    res.push([
      clusterCenter[0] + legLength * Math.cos(angle),
      clusterCenter[1] + legLength * Math.sin(angle),
    ]);
  }

  return res;
}

let hoverFeature;
/**
 * Style for convex hulls of clusters, activated on hover.
 * @param {Feature} cluster The cluster feature.
 * @return {Style} Polygon style for the convex hull of the cluster.
 */
function clusterHullStyle(featureStyles, cluster) {
  if (cluster !== hoverFeature) {
    return;
  }
  const originalFeatures = cluster.get('features');
  const points = originalFeatures.map((feature) =>
    feature.getGeometry().getCoordinates()
  );
  return new Style({
    geometry: new Polygon([monotoneChainConvexHull(points)]),
    fill: convexHullFill,
    stroke: convexHullStroke,
  });
}

function clusterStyle(featureStyles, feature) {
  const features = feature.get('features')
  const feature0 = features[0]

  if (features.length > 1) {
    return [
      clusterMemberStyle(featureStyles, feature0),
      new Style({
        text: new Text({
          text: features.length.toString(),
          fill: textFill,
          stroke: textStroke,
        }),
      }),
    ];
  }

  return clusterMemberStyle(featureStyles, feature0);
}

export {
  clusterMemberStyle,
  clusterCircleStyle,
  generatePointsCircle,
  clusterHullStyle,
  clusterStyle,
};
