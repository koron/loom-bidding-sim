import java.net.URL;
import java.net.HttpURLConnection;
import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.Random;

public class Client {
	private int sid;
	private int num;

	private volatile boolean closed;
	private volatile Result top;

	public static class Result {
		public int rid;
		public int score;
		public Result(int rid, int score) {
			this.rid = rid;
			this.score = score;
		}
	}

	public Client(int sid, int num) {
		this.sid = sid;
		this.num = num;
	}

	public Result run() {
		System.out.println(String.format("sid=%d num=%d", sid, num));
		final Object mu = new Object();

		for (int i = 0; i < num; i++) {
			final int rid = i;
			Fiber.execute(() -> {
				try {
					Result r = get(rid);
					System.out.println(String.format("complete rid=%d", rid));
					synchronized (mu) {
						put(r);
					}
				} catch (Exception e) {
					System.out.println(String.format("failed: %s", e));
				}
			});
		}

		try {
			Thread.sleep(100);
		} catch (InterruptedException e) {
		}
		System.out.println("closed");
		synchronized (mu) {
			closed = true;
		}

		return top;
	}

	public Result get(int rid) throws Exception {
		int score = 0;
		// get score from server.
		URL url = new URL(String.format("http://127.0.0.1:8080/?sid=%d&rid=%d", sid, rid));
		HttpURLConnection http = (HttpURLConnection) url.openConnection();
        http.setRequestMethod("GET");
        http.connect();
		if (http.getResponseCode() != 200) {
			throw new Exception("http failed");
		}
		BufferedReader r = new BufferedReader(new InputStreamReader(http.getInputStream()));
		String s = r.readLine();
		r.close();
		score = Integer.parseInt(s);
		return new Result(rid, score);
	}

	public void put(Result r) {
		if (closed || r == null) {
			return;
		}
		if (top == null || r.score > top.score) {
			top = r;
		}
	}

	public static void main(String[] args) {
		Random rnd = new Random();
		Client c = new Client(rnd.nextInt(1000), 100);
		Result r = c.run();
		if (r == null) {
			System.out.println("Client#run returns null");
			return;
		}
		System.out.println(String.format("id=%d score=%d", r.rid, r.score));
	}
}
