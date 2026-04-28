import { computed } from 'vue'

// 科技军事风格主题配置
export const useTechChartTheme = () => {
    // 科技蓝渐变色
    const gradientColors = [
        ['#00d4ff', '#0066ff'],
        ['#00ffff', '#00d4ff'],
        ['#39ff14', '#00ff88'],
        ['#ff073a', '#ff4466']
    ]

    // 基础颜色
    const colors = {
        primary: '#00d4ff',
        secondary: '#00ffff',
        success: '#39ff14',
        danger: '#ff073a',
        warning: '#ffcc00',
        bgDark: 'rgba(11, 26, 42, 0.9)',
        bgLight: 'rgba(0, 212, 255, 0.1)',
        text: '#ffffff',
        textMuted: '#99ccee',
        border: 'rgba(0, 212, 255, 0.3)'
    }

    // 折线图基础配置
    const lineChartOptions = computed(() => ({
        backgroundColor: 'transparent',
        grid: {
            top: 40,
            right: 30,
            bottom: 50,
            left: 60,
            containLabel: false
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            axisLine: {
                lineStyle: {
                    color: colors.border,
                    width: 1
                }
            },
            axisTick: {
                lineStyle: {
                    color: colors.primary,
                    width: 1
                }
            },
            axisLabel: {
                color: colors.textMuted,
                fontFamily: 'Orbitron, monospace',
                fontSize: 10,
                letterSpacing: 1
            },
            splitLine: {
                show: false
            }
        },
        yAxis: {
            type: 'value',
            axisLine: {
                lineStyle: {
                    color: colors.border,
                    width: 1
                }
            },
            axisTick: {
                lineStyle: {
                    color: colors.primary,
                    width: 1
                }
            },
            axisLabel: {
                color: colors.textMuted,
                fontFamily: 'Orbitron, monospace',
                fontSize: 10,
                letterSpacing: 1
            },
            splitLine: {
                lineStyle: {
                    color: 'rgba(0, 212, 255, 0.1)',
                    width: 1,
                    type: 'dashed'
                }
            }
        },
        tooltip: {
            trigger: 'axis',
            backgroundColor: colors.bgDark,
            borderColor: colors.primary,
            borderWidth: 1,
            textStyle: {
                color: colors.text,
                fontFamily: 'Orbitron, monospace',
                fontSize: 12
            },
            axisPointer: {
                type: 'cross',
                lineStyle: {
                    color: colors.primary,
                    width: 1,
                    type: 'dashed'
                },
                crossStyle: {
                    color: colors.primary,
                    width: 1
                }
            }
        },
        legend: {
            textStyle: {
                color: colors.text,
                fontFamily: 'Rajdhani, sans-serif',
                fontSize: 12
            },
            top: 10,
            itemWidth: 20,
            itemHeight: 10
        }
    }))

    // 创建荧光渐变折线系列配置
    const createGlowLineSeries = (
        name: string,
        data: number[],
        colorIndex: number = 0
    ) => {
        const color = gradientColors[colorIndex % gradientColors.length]
        return {
            name,
            type: 'line',
            smooth: true,
            symbol: 'circle',
            symbolSize: 6,
            lineStyle: {
                width: 3,
                color: {
                    type: 'linear',
                    x: 0,
                    y: 0,
                    x2: 1,
                    y2: 0,
                    colorStops: [
                        { offset: 0, color: color[0] },
                        { offset: 1, color: color[1] }
                    ]
                },
                shadowColor: color[0],
                shadowBlur: 10
            },
            itemStyle: {
                color: color[1],
                borderWidth: 2,
                borderColor: '#fff',
                shadowColor: color[0],
                shadowBlur: 15
            },
            areaStyle: {
                color: {
                    type: 'linear',
                    x: 0,
                    y: 0,
                    x2: 0,
                    y2: 1,
                    colorStops: [
                        { offset: 0, color: `${color[0]}40` },
                        { offset: 0.5, color: `${color[1]}20` },
                        { offset: 1, color: 'transparent' }
                    ]
                }
            },
            data,
            animationDuration: 2000,
            animationEasing: 'cubicOut'
        }
    }

    // 雷达图配置
    const radarChartOptions = computed(() => ({
        backgroundColor: 'transparent',
        radar: {
            indicator: [
                { name: '高度', max: 100 },
                { name: '速度', max: 50 },
                { name: '电量', max: 100 },
                { name: '信号', max: 100 },
                { name: '温度', max: 50 }
            ],
            shape: 'polygon',
            splitNumber: 4,
            axisName: {
                color: colors.textMuted,
                fontFamily: 'Rajdhani, sans-serif',
                fontSize: 11
            },
            splitLine: {
                lineStyle: {
                    color: colors.border,
                    width: 1
                }
            },
            splitArea: {
                areaStyle: {
                    color: ['transparent', 'transparent']
                }
            },
            axisLine: {
                lineStyle: {
                    color: colors.border,
                    width: 1
                }
            }
        },
        tooltip: {
            trigger: 'item',
            backgroundColor: colors.bgDark,
            borderColor: colors.primary,
            borderWidth: 1,
            textStyle: {
                color: colors.text,
                fontFamily: 'Orbitron, monospace'
            }
        },
        series: [{
            type: 'radar',
            data: [],
            lineStyle: {
                width: 2,
                color: colors.primary
            },
            areaStyle: {
                color: `${colors.primary}40`
            },
            itemStyle: {
                color: colors.primary
            }
        }]
    }))

    // 仪表盘配置
    const gaugeChartOptions = (value: number, name: string, max: number = 100) => ({
        backgroundColor: 'transparent',
        series: [{
            type: 'gauge',
            name,
            startAngle: 220,
            endAngle: -40,
            min: 0,
            max,
            splitNumber: 5,
            radius: '90%',
            axisLine: {
                lineStyle: {
                    width: 8,
                    color: [
                        [0.3, colors.danger],
                        [0.7, colors.warning],
                        [1, colors.success]
                    ]
                }
            },
            pointer: {
                itemStyle: {
                    color: colors.primary
                },
                width: 3,
                length: '60%'
            },
            axisTick: {
                distance: -15,
                length: 5,
                lineStyle: {
                    color: colors.primary,
                    width: 1
                }
            },
            splitLine: {
                distance: -20,
                length: 8,
                lineStyle: {
                    color: colors.primary,
                    width: 2
                }
            },
            axisLabel: {
                color: colors.textMuted,
                fontFamily: 'Orbitron, monospace',
                fontSize: 9,
                distance: -25
            },
            detail: {
                valueAnimation: true,
                fontSize: 20,
                fontFamily: 'Orbitron, monospace',
                fontWeight: 'bold',
                color: colors.primary,
                formatter: '{value}'
            },
            title: {
                color: colors.textMuted,
                fontFamily: 'Rajdhani, sans-serif',
                fontSize: 12,
                offsetCenter: [0, '70%']
            },
            data: [{ value, name }],
            animationDuration: 2000,
            animationEasing: 'cubicOut'
        }]
    })

    return {
        colors,
        gradientColors,
        lineChartOptions,
        radarChartOptions,
        gaugeChartOptions,
        createGlowLineSeries
    }
}
