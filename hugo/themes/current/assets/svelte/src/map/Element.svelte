<svelte:options customElement={{
	tag: 'dpb587-map',
	props: {
		center: { type: 'String', attribute: 'center' },
		zoom: { type: 'String', attribute: 'zoom' },
		sourcesJSON: { type: 'String', attribute: 'sources-json' },
		interactionMode: { type: 'String', attribute: 'interaction-mode' },
	}
}} />

<script>
  import Map from 'ol/Map'
  import GeoJSON from 'ol/format/GeoJSON';
  import View from 'ol/View'
  import TileLayer from 'ol/layer/Tile'
  import VectorLayer from 'ol/layer/Vector'
  import VectorSource from 'ol/source/Vector';
  import Cluster from 'ol/source/Cluster';
  import { fromLonLat } from 'ol/proj'
  import OSM from 'ol/source/OSM';
  import Overlay from 'ol/Overlay';
  import { defaults as controlDefaults } from 'ol/control'
  import { defaults as interactionDefaults } from 'ol/interaction'
  import { clusterHullStyle, clusterStyle, clusterCircleStyle } from './utils';
  import { defaultFeatureStyles, getFeatureStyle } from './utils.styles';
	import { onMount } from 'svelte';
  import { boundingExtent } from 'ol/extent';
  import { Icon, Style } from 'ol/style';
  
  //props 
  export let interactionMode = 'click';

  export let center = null;
  export let zoom = null;

  export let sourcesJSON = '[]';
  const sources = JSON.parse(sourcesJSON)
  
  let state = {};

  let domMap = null;

  onMount(() => {
    const featureStyles = {
      ...defaultFeatureStyles,
      Point: new Style({
        // requires HTMLImageElement from browser
        image: new Icon({
          src: 'https://img.icons8.com/plasticine/100/000000/place-marker.png',
          anchor: [0.45, 0.95],
          anchorXUnits: 'fraction',
          anchorYUnits: 'fraction',
          scale: 0.4,
          size: [ 100, 100],
        }),
        // image: new CircleStyle({
        //   fill: new Fill({
        //     color: 'rgba(50,103,227,0.9)',
        //   }),
        //   radius: 6,
        // }),
        // text: new Text({
        //   font: '12px Calibri,sans-serif',
        //   placement: 'line',
        //   fill: new Fill({
        //       color: '#000'
        //   }),
        //   stroke: new Stroke({
        //       color: '#fff',
        //       width: 2
        //   })
        // }),
      })
    }
    const localView = new View({
      center: center ? fromLonLat(center) : undefined,
      zoom,
    });

    const localOverlay = new Overlay({
      // element: domOverlay.current,
      autoPan: true,
      autoPanAnimation: {
        duration: 250,
      },
    });

    const localRasterLayer = new TileLayer({
      source: new OSM({
        // url: 'http://localhost:8089/=http%3A%2F%2Ftile.openstreetmap.org%2F{z}%2F{x}%2F{y}.png',
      }),
      preload: 2,
    });

    //

    const mapLayers = new Array(localRasterLayer);
    const allCoordinates = new Array();

    Promise.all(sources.map(layer => {
      if (layer.dataURL) {
        return fetch(layer.dataURL).then(async r => ({
          ...layer,
          data: await r.json(),
        }));
      }

      return layer;
    })).then(data => data.forEach((layer, idx) => {
      const features = new GeoJSON({ featureProjection: 'EPSG:3857' }).readFeatures(layer.data);
      const vectorSource = new VectorSource({
        features,
      })
      // const vectorLayer = new VectorLayer({
      //   source: vectorSource,
      //   style: getFeatureStyle,
      // })

      if (layer.cluster) {
        const clusterSource = new Cluster({
          distance: 35,
          source: vectorSource,
        });

        // Layer displaying the convex hull of the hovered cluster.
        const clusterHulls = new VectorLayer({
          source: clusterSource,
          style: v => clusterHullStyle(featureStyles, v),
        });

        // Layer displaying the clusters and individual features.
        const clusters = new VectorLayer({
          source: clusterSource,
          style: v => clusterStyle(featureStyles, v),
        });

        // Layer displaying the expanded view of overlapping cluster members.
        const clusterCircles = new VectorLayer({
          source: clusterSource,
          style: v => clusterCircleStyle(featureStyles, v),
        });

        // if (layers[idx].role != 'viewport') {
          mapLayers.push(clusterHulls, clusters, clusterCircles)
        // }
      } else {
        const vectorLayer = new VectorLayer({
          source: vectorSource,
          style: v => getFeatureStyle(featureStyles, v),
        });

        mapLayers.push(vectorLayer);
      }

      features.forEach((r) => {
        const geo = r.getGeometry()
        
        switch (geo.getType()) {
          case 'Point':
            allCoordinates.push(geo.getCoordinates())
            break
          case 'LineString':
            for (const set of geo.getCoordinates()) {
              allCoordinates.push(set)
            }
            break
          case 'Polygon':
            for (const set of geo.getCoordinates()) {
              allCoordinates.push(...set)
            }
            break
          case 'MultiLineString':
            for (const set of geo.getCoordinates()) {
              allCoordinates.push(...set.flat(0).map(v => [v[0], v[1]]));
            }
            break
          default:
            console.log(`missing behavior for ${geo.getType()}`)
        }
      })
    })).then(() => {
      if (allCoordinates.length > 0) {
        setTimeout(
          () => {
            localView.fit(
              boundingExtent(allCoordinates),
              {
                // padding: [ 32, 65, 30, 65 ],
                padding: [
                  25,
                  20,
                  interactionMode != 'button' ? 40 : 25,
                  interactionMode != 'button' ? 50 : 20,
                ],
                // duration: transition ? 500 : 0,
                // duration: 100,
                maxZoom: 14,
              },
            );
          },
          5
        );
      } else {
        localView.fit(
          boundingExtent([fromLonLat([ -90.199402, 38.627003 ])]),
          {
            maxZoom: 5,
          },
        )
      }

      const localMap = new Map({
        controls: interactionMode == 'button'
          ? []
          : controlDefaults({
            rotate: false,
          }),
        interactions: interactionMode == 'button'
          ? []
          : interactionDefaults({
            altShiftDragRotate: false,
            pinchRotate: false,
            mouseWheelZoom: interactionMode != 'click',
          }),
        layers: mapLayers,
        target: domMap,
        overlays: [localOverlay],
        view: localView,
      });

      setTimeout(
        () => {

          localMap.updateSize();
        },
        500
      )

      state.Map = localMap;
    });


    return () => {
      if (state.Map) {
        state.Map.setTarget(null);
        state.Map = null;
      }
    };
  });
</script>

<div style="position:relative;z-index:0;width:100%;height:100%" bind:this={domMap}></div>

<style lang="postcss">
  /* patched from node_modules/ol/ol.css since optimizer drops non-global rules
  ** prompt: `Within this file, update every CSS rule to be contained within a :global() tag`
  **/
  @import "ol.css";

  @reference "tailwindcss";

  :global(.ol-zoom), :global(.ol-full-screen) {
    @apply absolute bg-white/20 backdrop-blur-sm rounded-sm text-lg shadow -m-px;
  }

  :global(.ol-zoom) {
    @apply top-1.5 left-1.5;
  }

  :global(.ol-full-screen) {
    @apply top-1.5 right-1.5;
  }

  :global(.ol-zoom button:first-child),
  :global(.ol-full-screen button:first-child) {
    @apply rounded-t-sm;
  }

  :global(.ol-zoom button:last-child),
  :global(.ol-full-screen button:last-child) {
    @apply rounded-b-sm;
  }

  :global(.ol-zoom button),
  :global(.ol-full-screen button) {
    @apply m-px bg-white bg-white/90 flex items-center justify-center h-9 w-9 text-gray-700 hover:text-black hover:bg-white;
  }

  :global(.ol-attribution) {
    @apply absolute text-xs bg-white/20 backdrop-blur-sm bottom-0 right-0 p-px rounded-tl-sm text-gray-600 truncate max-w-full;
  }

  :global(.ol-attribution button) {
    @apply hidden;
  }

  :global(.ol-attribution ul) {
    @apply px-1 py-0.5 bg-white/90 space-x-1;
  }

  :global(.ol-attribution ul li) {
    @apply inline-block;
  }

  :global(.ol-overlay) {
    @apply absolute bottom-3 rounded w-64 bg-white/90 backdrop-blur-md border border-gray-400 shadow-xl text-center text-sm py-2 px-2 -translate-x-32;
  }
  /* 
  .ol-overlay {
    font-family: 'Lucida Grande', Verdana, Geneva, Lucida, Arial, Helvetica, sans-serif !important;
    font-size: 12px;
    position: absolute;
    background-color: white;
    -webkit-filter: drop-shadow(0 1px 4px rgba(0, 0, 0, 0.2));
    filter: drop-shadow(0 1px 4px rgba(0, 0, 0, 0.2));
    padding: 15px;
    border-radius: 10px;
    border: 1px solid #cccccc;
    bottom: 12px;
    left: -50px;
    min-width: 100px;
  } */

  :global(.ol-overlay:after),
  :global(.ol-overlay:before) {
    @apply absolute pointer-events-none w-0 h-0 border border-transparent top-full;
    content: " ";
  }

  :global(.ol-overlay:after) {
    @apply border-t-white left-32;
    border-width: 10px;
    margin-left: -10px;
  }

  :global(.ol-overlay:before) {
    @apply border-t-gray-400 left-32;
    border-width: 11px;
    margin-left: -11px;
  }

  :global(.ol-overlay-closer) {
    @apply hidden;
    text-decoration: none;
    position: absolute;
    top: 2px;
    right: 8px;
  }

  :global(.ol-overlay-closer:after) {
    content: "✖";
    color: #c3c3c3;
  }
</style>