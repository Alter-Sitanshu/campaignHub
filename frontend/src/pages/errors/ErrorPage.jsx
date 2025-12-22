import { useParams } from "react-router-dom";



const ErrorPage = () => {
    const { code } = useParams();
    return (
        <div className="page-container" style={{
            "width": "100vw",
            "height": "100vh",
            "display": "flex",
            "flexDirection": "column",
            "alignItems": "center",
            "justifyContent": "flex-start",
        }}>
            <h3>Sorry...</h3>
            <h1>Error <span className={`code-${code}`}>{ code }</span>, Retry Again</h1>
        </div>
    )
};

export default ErrorPage;