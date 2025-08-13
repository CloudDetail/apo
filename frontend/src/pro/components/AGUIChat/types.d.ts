declare module '*.module.scss' {
    const classes: { [key: string]: string }
    export default classes
}

declare module '*.scss' {
    const content: { [key: string]: string }
    export default content
}

declare module '*.svg' {
    const src: string
    export default src
} 