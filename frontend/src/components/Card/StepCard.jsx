const StepCard = ({step, head, desc}) => {
    return (
        <div className='steps-card'>
            <div className='step-num'>{step}</div>
            <h3>{head}</h3>
            <p>{desc}</p>
        </div>
    )
};

export default StepCard;