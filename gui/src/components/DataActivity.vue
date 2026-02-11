<template>
  <div class="mt-3">
    <div class="activity-chart-container">
      <div class="chart-header">
        <h5>Activity Timeline</h5>
        <div class="range-controls">
          <b-button-group size="sm" class="mr-2">
            <b-button
              v-for="preset in presets"
              :key="preset.label"
              :variant="activePreset === preset.label ? 'primary' : 'outline-secondary'"
              @click="applyPreset(preset)"
            >{{ preset.label }}</b-button>
          </b-button-group>
          <b-input-group size="sm" class="date-range-group">
            <b-form-input type="date" v-model="customStart" @change="applyCustomRange" />
            <b-input-group-text>~</b-input-group-text>
            <b-form-input type="date" v-model="customEnd" @change="applyCustomRange" />
          </b-input-group>
        </div>
      </div>
      <div class="chart-wrapper" ref="chartWrapper">
        <svg
          ref="svgChart"
          :viewBox="`0 0 ${width} ${height}`"
          class="activity-chart"
          @mousedown.prevent="onBrushStart"
          @mousemove.prevent="onBrushMove"
          @mouseup.prevent="onBrushEnd"
          @mouseleave="onBrushEnd"
        >
          <line x1="0" :y1="height - 20" :x2="width" :y2="height - 20" stroke="#ccc" />
          <text
            v-for="(label, i) in axisLabels"
            :key="'l'+i"
            :x="label.x"
            :y="height - 5"
            font-size="8"
            text-anchor="middle"
          >{{ label.text }}</text>
          <line x1="0" y1="0" x2="0" :y2="height - 20" stroke="#ccc" />
          <rect
            v-for="(point, index) in chartPoints"
            :key="'b'+index"
            :x="point.x"
            :y="point.y"
            :width="point.w"
            :height="point.h"
            fill="rgba(0, 123, 255, 0.6)"
            class="activity-bar"
          >
            <title>{{ point.title }}</title>
          </rect>
          <!-- Brush overlay -->
          <rect
            x="0" y="0" :width="width" :height="height - 20"
            fill="transparent"
            style="cursor: crosshair;"
          />
          <!-- Brush selection -->
          <rect
            v-if="brushing"
            :x="brushRect.x"
            y="0"
            :width="brushRect.w"
            :height="height - 20"
            fill="rgba(0, 123, 255, 0.15)"
            stroke="rgba(0, 123, 255, 0.6)"
            stroke-width="1"
            stroke-dasharray="4 2"
            pointer-events="none"
          />
        </svg>
        <div v-if="brushing" class="brush-tooltip" :style="brushTooltipStyle">
          {{ brushRangeLabel }}
        </div>
      </div>
      <div v-if="hasZoomed" class="zoom-hint mt-1">
        <b-button size="sm" variant="outline-secondary" @click="resetZoom">
          ↩ Reset to {{ zoomParentLabel }}
        </b-button>
        <small class="text-muted ml-2">Drag on chart to zoom in further</small>
      </div>
    </div>

    <b-table
      hover
      striped
      :items="activityItems"
      :fields="fields"
      responsive="sm"
      class="mt-4"
    >
      <template #cell(duration)="row">
        {{ formatDuration(row.item) }}
      </template>
    </b-table>
  </div>
</template>

<script>
import { get_field } from "./utils.js";
import moment from "moment";

export default {
  name: "DataActivity",
  props: {
    id: {
      type: String,
      required: true,
    },
    show: {
      type: Boolean,
      default: true,
    }
  },
  data() {
    return {
      activity: [],
      width: 800,
      height: 200,
      rangeStart: moment().subtract(30, 'days').startOf('day'),
      rangeEnd: moment().endOf('day'),
      activePreset: '30d',
      customStart: '',
      customEnd: '',
      brushing: false,
      brushStartX: 0,
      brushCurrentX: 0,
      zoomStack: [],
      presets: [
        { label: '24h', amount: 24, unit: 'hours' },
        { label: '7d', amount: 7, unit: 'days' },
        { label: '30d', amount: 30, unit: 'days' },
        { label: '90d', amount: 90, unit: 'days' },
      ],
      fields: [
        { key: "start", label: "Start", formatter: (v) => v ? moment(v).format("YYYY-MM-DD HH:mm:ss") : "-" },
        { key: "end", label: "End", formatter: (v) => v ? moment(v).format("YYYY-MM-DD HH:mm:ss") : "Still Active" },
        { key: "duration", label: "Duration" },
      ]
    };
  },
  computed: {
    hasZoomed() {
      return this.zoomStack.length > 0;
    },
    zoomParentLabel() {
      if (this.zoomStack.length === 0) return '';
      const prev = this.zoomStack[this.zoomStack.length - 1];
      return prev.label || (prev.start.format("MM/DD") + ' - ' + prev.end.format("MM/DD"));
    },
    brushRect() {
      const x1 = Math.min(this.brushStartX, this.brushCurrentX);
      const x2 = Math.max(this.brushStartX, this.brushCurrentX);
      return { x: x1, w: x2 - x1 };
    },
    brushRangeLabel() {
      if (!this.brushing) return '';
      const s = this.pxToTime(this.brushRect.x);
      const e = this.pxToTime(this.brushRect.x + this.brushRect.w);
      const fmt = this.rangeDurationHours <= 48 ? "MM/DD HH:mm" : "YYYY-MM-DD";
      return s.format(fmt) + ' ~ ' + e.format(fmt);
    },
    brushTooltipStyle() {
      const centerX = this.brushRect.x + this.brushRect.w / 2;
      const pct = (centerX / this.width) * 100;
      return { left: pct + '%', transform: 'translateX(-50%)' };
    },
    rangeDurationHours() {
      return this.rangeEnd.diff(this.rangeStart, 'hours', true);
    },
    slotConfig() {
      const hours = this.rangeDurationHours;
      if (hours <= 48) return { slots: Math.ceil(hours * 6), minutesPerSlot: 10, labelUnit: 'hour' };
      if (hours <= 168) return { slots: Math.ceil(hours), minutesPerSlot: 60, labelUnit: 'day' };
      return { slots: Math.ceil(hours / 24), minutesPerSlot: 1440, labelUnit: 'day' };
    },
    axisLabels() {
      const { slots, labelUnit } = this.slotConfig;
      const labels = [];
      const totalMinutes = this.rangeEnd.diff(this.rangeStart, 'minutes');
      if (labelUnit === 'hour') {
        const step = Math.max(1, Math.floor(slots / 24));
        for (let i = 0; i <= slots; i += step) {
          const t = moment(this.rangeStart).add(i * (totalMinutes / slots), 'minutes');
          labels.push({ x: (i / slots) * this.width, text: t.format("HH:mm") });
        }
      } else {
        const days = Math.ceil(totalMinutes / 1440);
        const step = Math.max(1, Math.floor(days / 15));
        for (let d = 0; d <= days; d += step) {
          const t = moment(this.rangeStart).add(d, 'days');
          labels.push({ x: (d / days) * this.width, text: t.format("MM/DD") });
        }
      }
      return labels;
    },
    activityItems() {
      const start = this.rangeStart;
      const end = this.rangeEnd;
      return [...this.activity]
        .filter(a => {
          const s = moment(a.start);
          const e = a.end ? moment(a.end) : moment();
          return e.isAfter(start) && s.isBefore(end);
        })
        .sort((a, b) => new Date(b.start) - new Date(a.start));
    },
    chartPoints() {
      if (this.activity.length === 0) return [];

      const now = moment();
      const startOfWindow = this.rangeStart;
      const endOfWindow = moment.min(this.rangeEnd, now);
      const { slots, minutesPerSlot } = this.slotConfig;
      const slotValues = new Array(slots).fill(0);

      this.activity.forEach(p => {
        const s = moment(p.start);
        const e = p.end ? moment(p.end) : now;
        if (e.isBefore(startOfWindow) || s.isAfter(endOfWindow)) return;

        const effectiveStart = moment.max(s, startOfWindow);
        const effectiveEnd = moment.min(e, endOfWindow);
        const startDiff = effectiveStart.diff(startOfWindow, 'minutes');
        const endDiff = effectiveEnd.diff(startOfWindow, 'minutes');

        for (let i = 0; i < slots; i++) {
          const slotStart = i * minutesPerSlot;
          const slotEnd = (i + 1) * minutesPerSlot;
          const overlapStart = Math.max(slotStart, startDiff);
          const overlapEnd = Math.min(slotEnd, endDiff);
          if (overlapEnd > overlapStart) {
            slotValues[i] += (overlapEnd - overlapStart);
          }
        }
      });

      const barWidth = this.width / slots;
      const points = [];
      slotValues.forEach((val, i) => {
        if (val > 0) {
          const h = (val / minutesPerSlot) * (this.height - 40);
          const slotTime = moment(startOfWindow).add(i * minutesPerSlot, 'minutes');
          const label = minutesPerSlot >= 1440
            ? `${slotTime.format("MM/DD")}: ${Math.round(val)} mins active`
            : minutesPerSlot >= 60
              ? `${slotTime.format("MM/DD HH:mm")}: ${Math.round(val)} mins active`
              : `${slotTime.format("HH:mm")}: ${Math.round(val)} mins active`;
          points.push({
            x: i * barWidth,
            y: this.height - 20 - h,
            w: Math.max(barWidth - 1, 1),
            h: h,
            title: label
          });
        }
      });
      return points;
    }
  },
  watch: {
    show() {
      if (this.show) {
        this.fetchData();
      }
    },
  },
  methods: {
    applyPreset(preset) {
      this.activePreset = preset.label;
      this.rangeStart = moment().subtract(preset.amount, preset.unit).startOf(preset.unit === 'hours' ? 'hour' : 'day');
      this.rangeEnd = moment().endOf('day');
      this.customStart = '';
      this.customEnd = '';
      this.zoomStack = [];
    },
    applyCustomRange() {
      if (this.customStart && this.customEnd) {
        const s = moment(this.customStart).startOf('day');
        const e = moment(this.customEnd).endOf('day');
        if (s.isValid() && e.isValid() && e.isAfter(s)) {
          this.zoomStack = [];
          this.rangeStart = s;
          this.rangeEnd = e;
          this.activePreset = '';
        }
      }
    },
    getSvgX(e) {
      const svg = this.$refs.svgChart;
      if (!svg) return 0;
      const rect = svg.getBoundingClientRect();
      const x = (e.clientX - rect.left) / rect.width * this.width;
      return Math.max(0, Math.min(this.width, x));
    },
    pxToTime(px) {
      const ratio = px / this.width;
      const totalMs = this.rangeEnd.diff(this.rangeStart);
      return moment(this.rangeStart).add(totalMs * ratio, 'ms');
    },
    onBrushStart(e) {
      this.brushing = true;
      this.brushStartX = this.getSvgX(e);
      this.brushCurrentX = this.brushStartX;
    },
    onBrushMove(e) {
      if (!this.brushing) return;
      this.brushCurrentX = this.getSvgX(e);
    },
    onBrushEnd() {
      if (!this.brushing) return;
      this.brushing = false;
      const minDrag = this.width * 0.02;
      if (Math.abs(this.brushCurrentX - this.brushStartX) < minDrag) return;

      const newStart = this.pxToTime(this.brushRect.x);
      const newEnd = this.pxToTime(this.brushRect.x + this.brushRect.w);

      this.zoomStack.push({
        start: moment(this.rangeStart),
        end: moment(this.rangeEnd),
        label: this.activePreset || (this.rangeStart.format("MM/DD") + '-' + this.rangeEnd.format("MM/DD"))
      });
      this.rangeStart = newStart;
      this.rangeEnd = newEnd;
      this.activePreset = '';
    },
    resetZoom() {
      if (this.zoomStack.length === 0) return;
      const prev = this.zoomStack.pop();
      this.rangeStart = prev.start;
      this.rangeEnd = prev.end;
      if (this.zoomStack.length === 0 && this.presets.some(p => p.label === prev.label)) {
        this.activePreset = prev.label;
      }
    },
    async fetchData() {
      try {
        const response = await get_field(this.id, "activity");
        this.activity = response || [];
      } catch (error) {
        console.error("Error fetching activity:", error);
      }
    },
    formatDuration(item) {
      if (!item.start) return "-";
      const start = moment(item.start);
      const end = item.end ? moment(item.end) : moment();
      const duration = moment.duration(end.diff(start));
      const parts = [];
      if (duration.days() > 0) parts.push(`${duration.days()}d`);
      if (duration.hours() > 0) parts.push(`${duration.hours()}h`);
      if (duration.minutes() > 0) parts.push(`${duration.minutes()}m`);
      parts.push(`${duration.seconds()}s`);
      return parts.join(" ");
    }
  },
  mounted() {
    this.fetchData();
  }
};
</script>

<style scoped>
.activity-chart-container {
  background: #fcfcfc;
  padding: 15px;
  border: 1px solid #eee;
  border-radius: 8px;
}
.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.chart-header h5 {
  margin: 0;
}
.range-controls {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}
.date-range-group {
  max-width: 320px;
}
.chart-wrapper {
  overflow-x: auto;
  position: relative;
  user-select: none;
}
.activity-chart {
  display: block;
  margin: 0 auto;
}
.activity-bar:hover {
  fill: #0056b3;
}
.brush-tooltip {
  position: absolute;
  top: -6px;
  padding: 2px 8px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  font-size: 11px;
  border-radius: 3px;
  white-space: nowrap;
  pointer-events: none;
}
.zoom-hint {
  display: flex;
  align-items: center;
}
</style>
