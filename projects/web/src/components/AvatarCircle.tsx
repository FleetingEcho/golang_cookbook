// ── 根据用户名生成稳定的颜色 ─────────────────────────────────────────────────

const AVATAR_COLORS = [
  '#2563eb', // blue
  '#0891b2', // cyan
  '#16a34a', // green
  '#ca8a04', // yellow
  '#dc2626', // red
  '#7c3aed', // purple
  '#db2777', // pink
  '#ea580c', // orange
  '#0d9488', // teal
  '#9333ea', // violet
];

export function userNameColor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}

// ── 圆形头像 SVG ────────────────────────────────────────────────────────────

interface Props {
  name: string;
  size?: number;
}

export function AvatarCircle({ name, size = 24 }: Props) {
  const color = userNameColor(name);
  const initial = name.charAt(0).toUpperCase();

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      aria-label={name}
      style={{ display: 'inline-block', verticalAlign: 'middle', flexShrink: 0 }}
    >
      <circle cx="12" cy="12" r="12" fill={color} />
      <text
        x="12"
        y="12"
        textAnchor="middle"
        dominantBaseline="central"
        fill="white"
        fontSize="12"
        fontFamily="system-ui, sans-serif"
        fontWeight="600"
      >
        {initial}
      </text>
    </svg>
  );
}
