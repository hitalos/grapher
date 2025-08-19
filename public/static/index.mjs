import { json } from 'd3'

import { Graph } from './graph.mjs'

document.querySelectorAll('svg.chart').forEach((el) => {
	const url = `${el.dataset.url}${window.location.search}`
	json(url).then((data) => {
		const graph = new Graph(el, data)
		const update = () => {
			json(url).then((data) => { graph.update(data) })
				.catch(console.error)
		}

		let updateInterval = setInterval(update, 10000)

		document.addEventListener('visibilitychange', () => {
			if (document.visibilityState !== 'visible') {
				clearInterval(updateInterval)
				return
			}
			updateInterval = setInterval(update, 10000)
		})

		window.addEventListener('resize', () => graph.resize())
	}).catch(console.error)
})