import { DynamicIcon, type IconName } from 'lucide-react/dynamic';


interface IconProps {
    name: IconName;
    size?: number;
    className?: string;
}


export const Icon = ({ name, size = 20, className }: IconProps) => {
    return (
        <DynamicIcon
            name={name}
            size={size}
            className={className}
            fallback={() => <div style={{ width: size, height: size }} />}
        />
    );
};