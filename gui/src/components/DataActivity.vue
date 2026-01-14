<template>
  <div class="mt-3">
    <div class="activity-chart-container">
      <h5>Activity Timeline (Last 24 Hours)</h5>
      <div class="chart-wrapper">
        <svg :viewBox="`0 0 ${width} ${height}`" class="activity-chart">
          <!-- Horizontal Axis (Hours) -->
          <line x1="0" :y1="height - 20" :x2="width" :y2="height - 20" stroke="#ccc" />
          <text v-for="h in 24" :key="h" :x="(width / 24) * (h - 1)" :y="height - 5" font-size="8">
            {{ formatHour(h - 1) }}
          </text>

          <!-- Vertical Axis (Active duration in seconds per slot) -->
          <line x1="0" y1="0" x2="0" :y2="height - 20" stroke="#ccc" />

          <!-- Bars/Line representing activity -->
          <rect
            v-for="(point, index) in chartPoints"
            :key="index"
            :x="point.x"
            :y="point.y"
            :width="point.w"
            :height="point.h"
            fill="rgba(0, 123, 255, 0.6)"
            class="activity-bar"
          >
            <title>{{ point.title }}</title>
          </rect>
        </svg>
      </div>
    </div>

    <!-- Original Table below for reference -->
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
      fields: [
        { key: "start", label: "Start", formatter: (v) => v ? moment(v).format("YYYY-MM-DD HH:mm:ss") : "-" },
        { key: "end", label: "End", formatter: (v) => v ? moment(v).format("YYYY-MM-DD HH:mm:ss") : "Still Active" },
        { key: "duration", label: "Duration" },
      ]
    };
  },
  computed: {
    activityItems() {
      return [...this.activity].sort((a,b) => new Date(b.start) - new Date(a.start));
    },
    chartPoints() {
      if (this.activity.length === 0) return [];
      
      const now = moment();
      const startOfWindow = moment().subtract(24, 'hours');
      const points = [];
      const slots = 144; // 10 minutes per slot (24 * 6)
      
      // Initialize slots
      const slotValues = new Array(slots).fill(0);
      
      this.activity.forEach(p => {
        const s = moment(p.start);
        const e = p.end ? moment(p.end) : now;
        
        if (e.isBefore(startOfWindow)) return;
        
        const effectiveStart = moment.max(s, startOfWindow);
        const effectiveEnd = moment.min(e, now);
        
        // Calculate which slots this duration occupies
        const startDiff = effectiveStart.diff(startOfWindow, 'minutes');
        const endDiff = effectiveEnd.diff(startOfWindow, 'minutes');
        
        const slotWidth = 1440 / slots; // minutes per slot
        
        for (let i = 0; i < slots; i++) {
          const slotStart = i * slotWidth;
          const slotEnd = (i + 1) * slotWidth;
          
          const overlapStart = Math.max(slotStart, startDiff);
          const overlapEnd = Math.min(slotEnd, endDiff);
          
          if (overlapEnd > overlapStart) {
            slotValues[i] += (overlapEnd - overlapStart);
          }
        }
      });

      const barWidth = this.width / slots;

      slotValues.forEach((val, i) => {
        if (val > 0) {
          const h = (val / (1440/slots)) * (this.height - 40); // % of slot occupied
          points.push({
            x: i * barWidth,
            y: this.height - 20 - h,
            w: barWidth - 1,
            h: h,
            title: `${Math.round(val)} mins active in this 10m slot`
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
    async fetchData() {
      try {
        const response = await get_field(this.id, "activity");
        this.activity = response || [];
      } catch (error) {
        console.error("Error fetching activity:", error);
      }
    },
    formatHour(h) {
      const time = moment().subtract(24 - h, 'hours');
      return time.format("H");
    },
    formatDuration(item) {
      if (!item.start) return "-";
      const start = moment(item.start);
      const end = item.end ? moment(item.end) : moment();
      const duration = moment.duration(end.diff(start));
      const parts = [];
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
.chart-wrapper {
  overflow-x: auto;
}
.activity-chart {
  display: block;
  margin: 0 auto;
}
.activity-bar:hover {
  fill: #0056b3;
}
</style>
