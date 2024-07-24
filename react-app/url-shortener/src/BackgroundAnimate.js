const BackgroundAnimate = () => {
  return (
    <ul className="background">
      {Array.from({ length: 13 }, (_, index) => (
        <li key={index}></li>
      ))}
    </ul>
  );
};

export default BackgroundAnimate;
