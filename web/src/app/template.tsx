/** 路由级模板:每次导航重新挂载,驱动 page-in 进场动画(reduced-motion 下 CSS 关闭)。 */
export default function Template({ children }: { children: React.ReactNode }) {
  return <div className="page-enter">{children}</div>;
}
