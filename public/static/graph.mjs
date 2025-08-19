import {
	axisLeft,
	axisBottom,
	interpolateRgbBasis,
	max,
	scaleBand,
	scaleLinear,
	scaleSequential,
	select,
} from 'd3'

const margin = { top: 40, right: 10, bottom: 60, left: 40 }

export class Graph {
	constructor(el, data) {
		this.svg = select(el)
		this.xValue = (d) => new Date(Date.parse(d.time))
		this.yValue = (d) => d.value
		this.title = this.svg.append('text')
			.classed('title', true)
			.attr('font-size', '2em')
			.attr('y', margin.top - 12)
			.attr('text-anchor', 'middle')
		this.g = this.svg.append('g')
		this.xAxis = this.g.append('g')
		this.yAxis = this.g.append('g')
		this.setDimensions()
			.update(data)
	}

	setDimensions() {
		this.width = this.svg.node().getBoundingClientRect().width
		this.height = this.svg.node().getBoundingClientRect().height
		this.innerHeight = this.height - margin.top - margin.bottom
		this.innerWidth = this.width - margin.left - margin.right
		this.title.attr('x', (this.width / 2))
		this.g.attr('transform', `translate(${margin.left},${margin.top})`)
		return this
	}

	setScales() {
		this.xScale = scaleBand()
			.domain(this.data.map(this.xValue))
			.range([0, this.innerWidth])
			.padding(0.1)
		this.yScale = scaleLinear()
			.domain([0, max(this.data, this.yValue)])
			.range([this.innerHeight, 0])
		this.colorScale = scaleSequential(interpolateRgbBasis(['#7777AA', 'navy']))
			.domain([0, max(this.data, this.yValue)])
		return this
	}

	setAxis() {
		this.yAxis.call(axisLeft(this.yScale))
		this.xAxis.call(axisBottom(this.xScale))
			.attr('transform', `translate(0, ${this.innerHeight})`)
			.selectAll('text')
			.attr('text-anchor', 'start')
			.attr('transform', 'translate(12, 8) rotate(90)')
			.text((d) => `${d.getFullYear()}-${d.getMonth()+1}-${d.getDate()}`)
			.append('title')
			.text((d) => d.toISOString())
		return this
	}

	resize() {
		this.setDimensions().setScales().setAxis()
		this.g.selectAll('rect')
			.attr('y', (d) => this.yScale(this.yValue(d)))
			.attr('x', (d) => this.xScale(this.xValue(d)))
			.attr('height', (d) => this.innerHeight - this.yScale(this.yValue(d)))
			.attr('width', this.xScale.bandwidth())
		return this
	}

	update(data) {
		this.data = data.results
		this.title.text(data.title)
		this.setScales().setAxis()
		const rects = this.g.selectAll('rect')
			.data(this.data, this.xValue)
		rects.transition().duration(500)
			.attr('y', (d) => this.yScale(this.yValue(d)))
			.attr('x', (d) => this.xScale(this.xValue(d)))
			.attr('height', (d) => this.innerHeight - this.yScale(this.yValue(d)))
			.attr('width', this.xScale.bandwidth())
		rects.exit()
			.transition().duration(200)
			.attr('height', 0)
			.attr('y', margin.top)
			.remove()
		const enters = rects.enter()
			.append('rect')
			.attr('x', (d) => this.xScale(this.xValue(d)))
			.attr('y', this.innerHeight)
			.attr('fill', (d) => this.colorScale(d.value))
			.attr('width', this.xScale.bandwidth())
			.attr('height', 0)
		enters.transition().duration(500)
			.delay((_, i) => i * 25)
			.attr('y', (d) => this.yScale(this.yValue(d)))
			.attr('height', (d) => this.innerHeight - this.yScale(this.yValue(d)))
		enters.append('title')
			.text((d) => d.attrs?.title ? d.attrs.title : d.value)
	}
}
