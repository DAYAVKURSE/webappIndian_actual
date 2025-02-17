import { useEffect } from 'react';
import styles from './Slider.module.scss';

export const Slider = ({ percent, setPercent, min, max, step }) => {
    const div = (max - min) / 100 || 1;

    useEffect(() => {
        const slider = document.querySelector('#slider');
        slider.style.setProperty('--value', `${slider.value / div}%`);

        slider.addEventListener('input', function() {
            this.style.setProperty('--value', `${this.value / div}%`);
        });
    }, [div]);

    const handlePercentChange = (value) => {
        let newValue = parseInt(value, 10);
    
        if (isNaN(newValue)) {
            setPercent(min || 0);
        } else if (newValue < min || 0) {
            setPercent(min || 0);
        } else if (newValue > max) {
            setPercent(max || 100);
        } else {
            setPercent(newValue);
        }
    };

    return (
        <input
            type="range"
            min={min}
            max={max}
            value={percent}
            step={step || 1}
            onChange={(e) => handlePercentChange(e.target.value)}
            className={styles.slider}
            id="slider"
        />
    );
};
