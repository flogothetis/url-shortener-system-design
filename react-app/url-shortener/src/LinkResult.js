import axios from "axios";
import { useEffect, useState } from "react";
import { redirect } from "react-router-dom";
import CopyToClipboard from "react-copy-to-clipboard";

const LinkResult = ({ inputValue }) => {
  const [shortenLink, setShortenLink] = useState("");
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(false);

  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await axios.post("http://localhost:8080/shorten", {
        originalUrl: inputValue,
      });

      const le = res.data.shortUrl;
      console.log("🚀 ~ fetchData ~ le:", le);

      setShortenLink(res.data.shortUrl);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (inputValue.length) {
      fetchData();
    }
  }, [inputValue]);

  const redirectToShortenedLink = async () => {
    try {
      setLoading(true);
      const res = await axios.get(`http://localhost:8080/${shortenLink}`, {
        mode: "no-cors",
        headers: {
          "Access-Control-Allow-Origin": "*",
          Accept: "application/json",
          "Content-Type": "application/json",
        },
      });

      console.log(
        "🚀 ~ redirectToShortenedLink ~ res:",
        res.request?.responseURL
      );
      // Redirect logic
      window.location.href = res.request?.responseURL;
      // Reload the window after redirecting
    } catch (err) {
      setError(err);
      //   window.location.reload();
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      {loading && <p className="noData">Loading...</p>}
      {error && <p className="noData">Something went wrong :(</p>}
      {shortenLink && (
        <div className="result">
          <p>{shortenLink}</p>
          <CopyToClipboard text={shortenLink} onCopy={() => setCopied(true)}>
            <button className={copied ? "copied" : ""}>
              Copy to Clipboard
            </button>
          </CopyToClipboard>
          <button onClick={redirectToShortenedLink}>Go to Link</button>
        </div>
      )}
    </>
  );
};

export default LinkResult;
